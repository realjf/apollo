# apollo

read from apolloconfig

### Example

```go
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

```
