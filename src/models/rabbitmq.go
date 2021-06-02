package models

import (
	"encoding/json"
	"fmt"
	"github.com/ichunt2019/go-rabbitmq/utils/rabbitmq"
	"push_service/src/config"
	"strings"

	//log "github.com/sirupsen/logrus"
	"push_service/src/core"
)

type RabbitMq struct {
	ID    int64
	Title string
}

type RecvPro struct {
}

func (t *RecvPro) Consumer(dataByte []byte) error {
	//处理数据 type =1 从数据库取值; type = 2 直接存入ES;
	//JsonStr := `{"titleCn":"中文","titleEn":"英文"}`
	var d config.PushEs
	//防止特殊字符导致报错
	//fmt.Println(string(httpBody))
	for i, ch := range dataByte {
		switch {
		case ch == '\r':
			dataByte[i] = ' '
		case ch == '\n':
			dataByte[i] = ' '
		case ch == '\t':
			dataByte[i] = ' '
		}
	}

	err := json.Unmarshal(dataByte, &d)
	if err != nil {
		fmt.Println("json err:", err)
	}

	//存入ES
	var EsModel EsModel
	EsModel.Init()
	for _, data := range d.Content {
		if primaryKey, ok := data["primaryKey"]; ok {
			c := make(map[string]string)
			for key, value := range data {
				if find := strings.Contains(key, "NestedEs"); find { //表示是嵌套字段
					fmt.Println("-----------")
					fmt.Println(value)
					EsModel.create(value, d.Name, primaryKey.(string))
				} else { //组装出没有嵌套字段的数组
					c[key] = value.(string)
				}
			}
			EsModel.create(c, d.Name, primaryKey.(string))
		}
	}
	return nil
}

func (t *RecvPro) FailAction(dataByte []byte) error {
	fmt.Println(string(dataByte))
	fmt.Println("任务处理失败了，我要进入db日志库了")
	fmt.Println("任务处理失败了，发送钉钉消息通知主人")
	return nil
}

func (RabbitMq) InitListenFirst(exName string, QuName string, ReKey string, ExType string, num int) {
	t := &RecvPro{}
	mqConf := core.Config.RabbitMqFirst
	dns := ""
	for _, conf := range mqConf {
		dns = conf.DataSourceName
	}

	rabbitmq.Recv(rabbitmq.QueueExchange{
		QuName,
		ReKey,
		exName,
		ExType,
		dns,
	}, t, num)
}

func (RabbitMq) InitListenSecond(exName string, QuName string, ReKey string, ExType string, num int) {
	t := &RecvPro{}
	mqConf := core.Config.RabbitMqFirst
	dns := ""
	for _, conf := range mqConf {
		dns = conf.DataSourceName
	}

	rabbitmq.Recv(rabbitmq.QueueExchange{
		QuName,
		ReKey,
		exName,
		ExType,
		dns,
	}, t, num)
}
