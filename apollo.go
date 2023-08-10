// #############################################################################
// # File: apollo.go                                                           #
// # Project: apollo                                                           #
// # Created Date: 2023/08/10 17:59:31                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 20:54:49                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package apollo

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/constant"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/extension"
	"github.com/go-kratos/kratos/v2/encoding"
	_ "github.com/go-kratos/kratos/v2/encoding/form"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
	_ "github.com/go-kratos/kratos/v2/encoding/proto"
	_ "github.com/go-kratos/kratos/v2/encoding/xml"
	_ "github.com/go-kratos/kratos/v2/encoding/yaml"
)

var _ Apollo = (*apollo)(nil)

func New() Apollo {
	apollo := &apollo{}
	return apollo
}

type Apollo interface {
	Init(cfg *config.AppConfig, isOriginConfig bool) (err error)
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}

type apollo struct {
	listener       *changeListener
	client         agollo.Client
	appConfig      *config.AppConfig
	isOriginConfig bool
}

func (a *apollo) Init(cfg *config.AppConfig, isOriginConfig bool) (err error) {
	a.appConfig = cfg
	a.isOriginConfig = isOriginConfig
	if isOriginConfig {
		// 使用原始配置文件没有经过解析处理
		extension.AddFormatParser(constant.JSON, &jsonExtParser{})
		extension.AddFormatParser(constant.YAML, &yamlExtParser{})
		extension.AddFormatParser(constant.YML, &yamlExtParser{})
	}

	a.client, err = agollo.StartWithConfig(func() (*config.AppConfig, error) {
		fmt.Printf("%#v\n", a.appConfig)
		return a.appConfig, nil
	})
	if err != nil {
		return
	}

	_, err = a.Load()
	if err != nil {
		return
	}

	runtime.SetFinalizer(a, stopApollo)

	return nil
}

func (a *apollo) Load() ([]*KeyValue, error) {
	kvs := make([]*KeyValue, 0)
	namespaces := strings.Split(a.appConfig.NamespaceName, ",")

	for _, ns := range namespaces {
		if !a.isOriginConfig {
			kv, err := a.getConfig(ns)
			if err != nil {
				log.Errorf("apollo get config failed，err:%v", err)
				continue
			}
			kvs = append(kvs, kv)
			continue
		}
		if strings.Contains(ns, ".") && !strings.HasSuffix(ns, "."+properties) &&
			(format(ns) == yaml || format(ns) == yml || format(ns) == json) {
			kv, err := a.getOriginConfig(ns)
			if err != nil {
				log.Errorf("apollo get config failed，err:%v", err)
				continue
			}
			kvs = append(kvs, kv)
			continue
		}
		kv, err := a.getConfig(ns)
		if err != nil {
			log.Errorf("apollo get config failed，err:%v", err)
			continue
		}
		kvs = append(kvs, kv)
	}
	return kvs, nil
}

func (e *apollo) getConfig(namespace string) (*KeyValue, error) {
	next := map[string]interface{}{}
	e.client.GetConfigCache(namespace).Range(func(key, value interface{}) bool {
		// all values are out properties format
		resolve(genKey(namespace, key.(string)), value, next)
		return true
	})
	f := format(namespace)
	codec := encoding.GetCodec(f)
	val, err := codec.Marshal(next)
	if err != nil {
		return nil, err
	}
	return &KeyValue{
		Key:    namespace,
		Value:  val,
		Format: f,
	}, nil
}

func (a *apollo) getOriginConfig(namespace string) (*KeyValue, error) {
	value, err := a.client.GetConfigCache(namespace).Get("content")
	if err != nil {
		return nil, err
	}
	// serialize the namespace content KeyValue into bytes.
	return &KeyValue{
		Key:    namespace,
		Value:  []byte(value.(string)),
		Format: format(namespace),
	}, nil
}

// Watch 监听配置变更
func (a *apollo) Watch() (Watcher, error) {
	w, err := newWatcher(a)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func stopApollo(a *apollo) {
	a.client.RemoveChangeListener(a.listener)
	a.client.Close()
}

/**************************************** from go-kratos/kratos project ********************************************************/
// resolve convert kv pair into one map[string]interface{} by split key into different
// map level. such as: app.name = "application" => map[app][name] = "application"
func resolve(key string, value interface{}, target map[string]interface{}) {
	// expand key "aaa.bbb" into map[aaa]map[bbb]interface{}
	keys := strings.Split(key, ".")
	last := len(keys) - 1
	cursor := target

	for i, k := range keys {
		if i == last {
			cursor[k] = value
			break
		}

		// not the last key, be deeper
		v, ok := cursor[k]
		if !ok {
			// create a new map
			deeper := make(map[string]interface{})
			cursor[k] = deeper
			cursor = deeper
			continue
		}

		// current exists, then check existing value type, if it's not map
		// that means duplicate keys, and at least one is not map instance.
		if cursor, ok = v.(map[string]interface{}); !ok {
			log.Warnf("duplicate key: %v\n", strings.Join(keys[:i+1], "."))
			break
		}
	}
}

// genKey got the key of config.KeyValue pair.
// eg: namespace.ext with subKey got namespace.subKey
func genKey(namespace, sub string) string {
	arr := strings.Split(namespace, ".")
	if len(arr) == 1 {
		if namespace == "" {
			return sub
		}

		return namespace + "." + sub
	}

	suffix := arr[len(arr)-1]
	_, ok := formats[suffix]
	if ok {
		return strings.Join(arr[:len(arr)-1], ".") + "." + sub
	}

	return namespace + "." + sub
}

/*****************************************************************************************************************************/
