// #############################################################################
// # File: watcher.go                                                          #
// # Project: apollo                                                           #
// # Created Date: 2023/08/10 19:25:13                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 20:42:43                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package apollo

import (
	"context"
	"strings"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/log"
)

type changeListener struct {
	in     chan<- []*KeyValue
	apollo *apollo
}

func newChangeListener(in chan<- []*KeyValue, apollo *apollo) *changeListener {
	return &changeListener{
		in:     in,
		apollo: apollo,
	}
}

func (c *changeListener) OnChange(changeEvent *storage.ChangeEvent) {
	kv := make([]*KeyValue, 0)
	namespace := changeEvent.Namespace
	if strings.Contains(namespace, ".") && !strings.HasSuffix(namespace, "."+properties) &&
		(format(namespace) == yaml || format(namespace) == yml || format(namespace) == json) {
		value, err := c.apollo.client.GetConfigCache(namespace).Get("content")
		if err != nil {
			log.Warnw("apollo get config failed", "err", err)
			return
		}
		kv = append(kv, &KeyValue{
			Key:    namespace,
			Value:  []byte(value.(string)),
			Format: format(namespace),
		})

	} else {
		next := make(map[string]interface{})

		for key, change := range changeEvent.Changes {
			resolve(genKey(namespace, key), change.NewValue, next)
		}

		f := format(namespace)
		codec := encoding.GetCodec(f)
		val, err := codec.Marshal(next)
		if err != nil {
			log.Warnf("apollo could not handle namespace %s: %v", namespace, err)
			return
		}
		kv = append(kv, &KeyValue{
			Key:    namespace,
			Value:  val,
			Format: f,
		})
	}

	if len(kv) == 0 {
		return
	}
	c.in <- kv
}

func (c *changeListener) OnNewestChange(event *storage.FullChangeEvent) {

}

type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}

type watcher struct {
	out <-chan []*KeyValue

	ctx        context.Context
	cancelFunc func()
}

func newWatcher(a *apollo) (Watcher, error) {
	changeChan := make(chan []*KeyValue)
	listener := newChangeListener(changeChan, a)
	a.client.AddChangeListener(listener)

	ctx, cancel := context.WithCancel(context.Background())
	return &watcher{
		out: changeChan,

		ctx: ctx,
		cancelFunc: func() {
			a.client.RemoveChangeListener(listener)
			cancel()
		},
	}, nil
}

// Next will be blocked until the Stop method is called
func (w *watcher) Next() ([]*KeyValue, error) {
	select {
	case kv := <-w.out:
		return kv, nil
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	}
}

func (w *watcher) Stop() error {
	if w.cancelFunc != nil {
		w.cancelFunc()
	}
	return nil
}
