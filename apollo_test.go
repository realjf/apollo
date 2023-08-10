// #############################################################################
// # File: apollo_test.go                                                      #
// # Project: apollo                                                           #
// # Created Date: 2023/08/10 18:01:52                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 20:53:38                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package apollo_test

import (
	"testing"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/realjf/apollo"
)

func TestApollo(t *testing.T) {
	cfg := &config.AppConfig{
		AppID:            "test",
		Cluster:          "dev",
		IP:               "http://localhost:8080",
		NamespaceName:    "application",
		IsBackupConfig:   true,
		Secret:           "",
		BackupConfigPath: ".",
	}

	apolloReader := apollo.New()
	if err := apolloReader.Init(cfg, true); err != nil {
		t.Fatal(err)
		return
	}

	watcher, err := apolloReader.Watch()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer watcher.Stop()

	for {
		kvs, err := watcher.Next()
		if err != nil {
			t.Fatal(err)
			continue
		}
		for _, v := range kvs {
			t.Logf("%v: %s", v.Key, v.Value)
		}
	}
}
