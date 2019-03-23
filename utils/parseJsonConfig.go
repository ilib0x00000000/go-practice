package utils


import (
	"io/ioutil"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

/**
	解析json配置文件， 返回一个map

	如果直接使用 json.Unmarshal 解析json数据时， 不指定接数据的格式，所有的int都默认解析为 float64

	2种解析模式：
		1. 直接解析
		2. 使用编码/解码形式  decode/encode  ---> 提前定义好 struct{}
 */
func parseJsonConfig(filename string) (map[string]interface{}, error) {
	var temp interface{}

	// 读取配置文件
	context, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// json parse
	err = json.Unmarshal(context, &temp)
	if err != nil {
		return nil, err
	}

	//fmt.Println(temp)
	//fmt.Println(reflect.TypeOf(temp))

	if result, ok := temp.(map[string]interface{}); ok {
		return result, nil
	}

	return nil, errors.New("parse json result error")
}

/**
	Json 直接解析到 struct时， struct的字段名必须是外部包可见的（首字母大写）
 */
type ServerConfig struct {
	Version string    `json:"version"`
	Ptype   int       `json:"type"`
	Name    string    `json:"name"`
	Age     int       `json:"age"`
	Have    float64   `json:"have"`
}


func decodeJsonConfig(filename string) (*ServerConfig, error) {
	config := ServerConfig{}

	contexts, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contexts, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}


