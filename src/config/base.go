package config

//推送数据结构
type PushEs struct {
	Content []map[string]interface{} `json:"content"` //内容
	Name    string              `json:"name"`    //索引
	Type    int                 `json:"type"`    //类型
}

//查询时的数据结构
type SearchEs struct {
	Condition  []map[string]string `json:"condition"`   //搜索参数
	OrderType  int                 `json:"order_type"`  // 1= 升序 2=降序
	OrderField string              `json:"order_field"` //排序字段
	Name       string              `json:"name"`        //索引
	Type       int                 `json:"type"`        //类型
	Limit      int                 `json:"limit"`       //每页数量
	Page       int                 `json:"page"`        //页数
}

//创建mapping
type AddMapping struct {
	Index   string `json:"index"`
	Mapping string `json:"mapping"`
}

//监听MQ
type ListenMq struct {
	ExName string `json:"ex_name"`
	QuName string `json:"qu_name"`
	ReKey  string `json:"re_key"`
	ExType string `json:"ex_type"`
	Num    int  `json:"num"`
}

//返回数据结构
type Result struct {
	PrimaryKey string `json:"primaryKey"`
}

//返回的组装参数
type Data struct {
	Total   int64    `json:"total"`
	Size    int64    `json:"size"`
	Page    int64    `json:"page"`
	Records []string `json:"records"`
}
