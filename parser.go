// #############################################################################
// # File: parser.go                                                           #
// # Project: apollo                                                           #
// # Created Date: 2023/08/10 18:23:44                                         #
// # Author: realjf                                                            #
// # -----                                                                     #
// # Last Modified: 2023/08/10 19:48:25                                        #
// # Modified By: realjf                                                       #
// # -----                                                                     #
// # Copyright (c) 2023                                                        #
// #############################################################################
package apollo

import "strings"

/**************************************** from go-kratos/kratos project ********************************************************/
const (
	yaml       = "yaml"
	yml        = "yml"
	json       = "json"
	properties = "properties"
)

var formats map[string]struct{}

func init() {
	formats = make(map[string]struct{})

	formats[yaml] = struct{}{}
	formats[yml] = struct{}{}
	formats[json] = struct{}{}
	formats[properties] = struct{}{}
}

type jsonExtParser struct{}

func (parser jsonExtParser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"content": configContent}, nil
}

type yamlExtParser struct{}

func (parser yamlExtParser) Parse(configContent interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"content": configContent}, nil
}

func format(namespace string) string {
	arr := strings.Split(namespace, ".")
	suffix := arr[len(arr)-1]
	if len(arr) <= 1 || suffix == properties {
		return json
	}
	if _, ok := formats[suffix]; !ok {
		// fallback
		return json
	}

	return suffix
}

/*****************************************************************************************************************************/
