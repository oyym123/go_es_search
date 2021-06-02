package controllers

import (
	"encoding/json"
	"fmt"
	_ "github.com/ichunt2019/go-rabbitmq/utils/rabbitmq"
	"net/http"
	"push_service/src/config"
	"push_service/src/models"
)

type GetMsgController struct {
	BaseController
}


/**
 * 初始化对应的ES映射表 【已存在的不会重新创建】
 */
func (c *GetMsgController) CreateMapping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		c.sendError(w, 400, "method not allowed!")
		return
	}
	var es models.EsModel
	es.Init()
	res := es.CreateMapping(w, r)
	c.sendOk(w, res)
}

/**
 * 监听队列
 */
func (c *GetMsgController) MqConn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		c.sendError(w, 400, "method not allowed!")
		return
	}

	var conf config.ListenMq

	decoder := json.NewDecoder(r.Body)
	fmt.Println("-------------")
	if err := decoder.Decode(&conf); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.sendError(w, 400, "不是正确的参数!")
	}
	fmt.Println(conf)

	var mq models.RabbitMq
	mq.InitListenFirst(conf.ExName, conf.QuName, conf.ReKey, conf.ExType, conf.Num)
	c.sendOk(w, "监听成功!")
}
