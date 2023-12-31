// #############################################################################
// # File: test.go                                                             #
// # Project: test                                                             #
// # Created Date: 2023/08/10 20:57:12                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 20:59:24                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package main

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/realjf/apollo"
)

func main() {
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
		fmt.Errorf(err.Error())
		return
	}

	watcher, err := apolloReader.Watch()
	if err != nil {
		fmt.Errorf(err.Error())
		return
	}
	defer watcher.Stop()

	for {
		kvs, err := watcher.Next()
		if err != nil {
			fmt.Errorf(err.Error())
			continue
		}
		for _, v := range kvs {
			fmt.Printf("%v: %s", v.Key, v.Value)
		}
	}
}
