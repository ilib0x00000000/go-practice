package utils


/**
	解析XML数据
		1. XML数据格式是固定的   直接使用 xml.Unmarshal
		2. XML的节点是动态的，比如微信支付返回的XML数据，可能是 <coupon_id_1></coupon_id_1>
			也可能是 <coupon_id_10></coupon_id_10>...，tag是不确定的，只能挨个读取XML元素

	参考： https://blog.csdn.net/yuanjize1996/article/details/84703964
 */


import (
	"io/ioutil"
	"encoding/xml"
	"fmt"
	"strings"
	"io"
	"reflect"
)


type XMLConfigData struct {
	//XMLName xml.Name    `xml:"config"`
	Name    string      `xml:"name"`
	Age     int         `xml:"age"`
	Have    float64     `xml:"have"`
	Address string      `xml:"address"`
	Phone   string      `xml:"phone"`
}


/**
	解析固定XML配置文件     ---  对应到struct中
 */
func parseXmlConfig(filename string) (*XMLConfigData, error) {
	var result XMLConfigData

	contexts, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(contexts, &result)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	return &result, nil
}


/**
	解析动态的XML数据
 */
func parseDynamicXML(filename string) (map[string]interface{}, error) {
	contexts, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// []byte  to string
	xmlStr := string(contexts)

	key := ""
	var result = make(map[string]interface{})

	reader_p := strings.NewReader(xmlStr)      // 构建一个*reader
	decoder := xml.NewDecoder(reader_p)        // 构建一个decoder

	for {
		token, err := decoder.Token()         // 读取一个标签 或者 文本 内容

		if err == io.EOF {
			return result, nil                // 解析结束
		}

		if err != nil {
			return nil, err
		}

		// token 可以是3种类型   StartElement起始标签，EndElement结束标签，CharData文本内容
		switch tp := token.(type) {
		case xml.StartElement:
			se := xml.StartElement(tp)   // 强制类型转换
			if se.Name.Local != "xml" {
				key = se.Name.Local
			}

			if len(se.Attr) != 0 {
				for _, v := range se.Attr {
					fmt.Println("attr of " + se.Name.Local +  " key: ", v.Name.Local, "   value: ", v.Value)
				}
				// fmt.Println("Attrs: ", se.Attr )     // 有的XML节点数据带有attr
				//if len(se.Attr) != 0 {
				//	for _, v := range se.Attr {
				//		fmt.Println("attr...  name: ", v.Name, "   value: ", v.Value)
				//	}
				//}
			}
		case xml.CharData:
			cd := xml.CharData(tp)
			data := strings.TrimSpace(string(cd))

			if len(data) != 0 {
				result[key] = data              // 如果数据存在，就存到map中
			}
		//case xml.EndElement:                  // 可以不用
		//	ee := xml.EndElement(tp)
		//
		//	if ee.Name.Local == "xml" {
		//		return result, nil             // 判断是不是 </xml>  如果是，返回
		//	}
		}
	}

	return result, nil
}

