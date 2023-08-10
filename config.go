// #############################################################################
// # File: config.go                                                           #
// # Project: apollo                                                           #
// # Created Date: 2023/08/10 19:14:35                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 20:41:44                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package apollo

type KeyValue struct {
	Key    string // namespace
	Value  []byte // value
	Format string // format json,yaml,yml,etc.
}

type ApolloConfig struct {
	AppID            string
	Cluster          string
	IP               string
	NamespaceName    string
	IsBackupConfig   bool
	Secret           string
	BackupConfigPath string
}
