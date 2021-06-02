package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	//"gopkg.in/olivere/elastic.v6"
	"log"
	"net/http"
	"os"
	"push_service/src/config"
	"reflect"
	"time"

	//"reflect"
)

var client *elastic.Client
var host = "http://192.168.44.8:9200/"

type EsModel struct {
	ID    int64
	Title string
}

type Employee struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	About     string   `json:"about"`
	Interests []string `json:"interests"`
}

//初始化
func (EsModel) Init() {
	errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	var err error
	client, err = elastic.NewClient(elastic.SetErrorLog(errorlog), elastic.SetURL(host))
	if err != nil {
		panic(err)
	}
	info, code, err := client.Ping(host).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)
}

/*下面是简单的CURD*/

//创建index
func CreateIndex(index string, mapping string) {

	exists, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		panic(err)
	}

	if !exists {
		// 如果不存在，就创建
		createIndex, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			println()
		}
	}
}

//新增mapping
func (EsModel) CreateMapping(w http.ResponseWriter, r *http.Request) interface{} {
	var data config.AddMapping
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return "不是正确的参数!"
	}

	//判断有没有index
	CreateIndex(data.Index, data.Mapping)
	_, err := client.PutMapping().Index(data.Index).BodyString(data.Mapping).Do(ctx)
	if err != nil {
		return err
	}

	return "创建成功!"
}

//创建Doc
func (EsModel) create(string2 interface{}, index string, primaryKey string) {
	fmt.Println("-----------------")
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(timeStr)
	fmt.Println(string2)
	//通过id查找 查询是否有这个值 有则修改 没有则新增
	_, err := client.Get().Index(index).Id(primaryKey).Do(context.Background())
	if err != nil { //新增
		res, err := client.Index().
			Index(index).
			Id(primaryKey).
			BodyJson(string2).
			Do(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("creates success %s\n", res.Result)
	} else { //修改
		res, err := client.Update().
			Index(index).

			Id(primaryKey).
			Doc(string2).
			Do(context.Background())
		if err != nil {
			println(err.Error())
		}
		if res != nil {
			fmt.Printf("update success %s\n", res.Result)
		}
	}
}

func (EsModel) Search(w http.ResponseWriter, r *http.Request) interface{} {
	var data config.SearchEs
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return ""
	}

	//解析查询因子
	query := elastic.NewBoolQuery()
	for _, d := range data.Condition {
		if searchType, ok := d["search"]; ok {
			if field, ok := d["field"]; ok {
				fmt.Println(d)
				fmt.Println(field)
				fmt.Println(searchType)
				switch searchType {
				case "=": //精准匹配
					query = query.Must(elastic.NewTermQuery(field, d["value"]))
					break
				case ">":
					query.Filter(elastic.NewRangeQuery(field).Gt(d["value"]))
					break
				case ">=":
					query.Filter(elastic.NewRangeQuery(field).Gte(d["value"]))
					break
				case "<":
					query.Filter(elastic.NewRangeQuery(field).Lt(d["value"]))
					break
				case "<=":
					query.Filter(elastic.NewRangeQuery(field).Lte(d["value"]))
					break
				case "!=":
					query.MustNot(elastic.NewTermQuery(field, d["value"]))
					break
				case "like":
					query.Must(elastic.NewMatchPhraseQuery(field, d["value"]))
					break
				case "not_like":
					query.MustNot(elastic.NewMatchPhraseQuery(field, d["value"]))
					break
				case "in":
					var value []interface{}
					_ = json.Unmarshal([]byte(d["value"]), &value)
					query.Must(elastic.NewTermsQuery(field, value...))
					break
				case "not_in":
					var value []interface{}
					_ = json.Unmarshal([]byte(d["value"]), &value)
					query.MustNot(elastic.NewTermsQuery(field, value...))
					break
				case "between_left": // <= X <
					query.Filter(elastic.NewRangeQuery(field).Gte(d["value_l"]))
					query.Filter(elastic.NewRangeQuery(field).Lt(d["value_r"]))
					break
				case "between_right": // < X <=
					query.Filter(elastic.NewRangeQuery(field).Gt(d["value_l"]))
					query.Filter(elastic.NewRangeQuery(field).Lte(d["value_r"]))
					break
				case "between": // <= X <=
					query.Filter(elastic.NewRangeQuery(field).Gte(d["value_l"]))
					query.Filter(elastic.NewRangeQuery(field).Lte(d["value_r"]))
					break
				case "should_between_right":
					query.Should(elastic.NewRangeQuery(field).Gt(d["value_l"]))
					query.Should(elastic.NewRangeQuery(field).Lte(d["value_r"]))
					break
				case "should_between_left":
					query.Should(elastic.NewRangeQuery(field).Gte(d["value_l"]))
					query.Should(elastic.NewRangeQuery(field).Lt(d["value_r"]))
					break
				case "should_between":
					query.Should(elastic.NewRangeQuery(field).Gte(d["value_l"]))
					query.Should(elastic.NewRangeQuery(field).Lte(d["value_r"]))
					break
				case "should_gt":
					query.Should(elastic.NewRangeQuery(field).Gt(d["value"]))
					break
				case "should_lt":
					query.Should(elastic.NewRangeQuery(field).Lt(d["value"]))
					break
				case "should_gte":
					query.Should(elastic.NewRangeQuery(field).Gte(d["value"]))
					break
				case "should_lte":
					query.Should(elastic.NewRangeQuery(field).Lte(d["value"]))
					break
				case "json_search": //Nested 嵌套搜索
					fmt.Println("---333-----")
					fmt.Println(d)
					if path, ok := d["path"]; ok {
						fmt.Println("---555-----")
						if search, ok := d["path_search"]; ok {
							fmt.Println(search)
							fmt.Println(path)
							fmt.Println("--------")
							//默认精准匹配
							switch search {
							case "should=": //用于嵌套的in查询
								fmt.Println(d["value"])
								q := elastic.NewTermQuery(path, d["value"])
								query.Should(elastic.NewNestedQuery(field, q))
								break
							case "=":
								q := elastic.NewTermQuery(path, d["value"])
								query.Must(elastic.NewNestedQuery(field, q))
								break
							case ">":
								q := elastic.NewRangeQuery(path).Gt(d["value"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case ">=":
								q := elastic.NewRangeQuery(path).Gte(d["value"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case "<":
								q := elastic.NewRangeQuery(path).Lt(d["value"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case "<=":
								q := elastic.NewRangeQuery(path).Lte(d["value"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case "!=":
								q := elastic.NewTermQuery(path, d["value"])
								query.MustNot(elastic.NewNestedQuery(field, q))
								break
							case "like":
								q := elastic.NewMatchPhraseQuery(path, d["value"])
								query.Must(elastic.NewNestedQuery(field, q))
								break
							case "not_like":
								q := elastic.NewMatchPhraseQuery(path, d["value"])
								query.MustNot(elastic.NewNestedQuery(field, q))
								break
							case "in":
								var value []interface{}
								_ = json.Unmarshal([]byte(d["value"]), &value)
								q := elastic.NewTermsQuery(path, value...)
								query.Must(elastic.NewNestedQuery(field, q))
								break
							case "not_in":
								var value []interface{}
								_ = json.Unmarshal([]byte(d["value"]), &value)
								q := elastic.NewTermsQuery(path, value...)
								query.MustNot(elastic.NewNestedQuery(field, q))

								break
							case "between_left": // <= X <
								q := elastic.NewRangeQuery(path).Gte(d["value_l"])
								query.Filter(elastic.NewNestedQuery(field, q))
								q = elastic.NewRangeQuery(path).Lt(d["value_r"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case "between_right": // < X <=
								q := elastic.NewRangeQuery(path).Gt(d["value_l"])
								query.Filter(elastic.NewNestedQuery(field, q))
								q = elastic.NewRangeQuery(path).Lte(d["value_r"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							case "between": // <= X <=
								q := elastic.NewRangeQuery(path).Gte(d["value_l"])
								query.Filter(elastic.NewNestedQuery(field, q))
								q = elastic.NewRangeQuery(path).Lte(d["value_r"])
								query.Filter(elastic.NewNestedQuery(field, q))
								break
							}
						}
					}
					break
				}
			}
		}
	}

	var size = data.Limit
	var page = data.Page

	if size < 0 || page < 1 {
		return "分页设置错误!"
	}

	//只取主键
	fsc := elastic.NewFetchSourceContext(true).Include("primaryKey")
	res, err := client.Search(data.Name).
		Query(query).
		FetchSourceContext(fsc).
		Size(size).
		Preference("primary_first").
		From((page - 1) * size).
		TrackTotalHits(true).
		Do(context.Background())

	if err != nil {
		panic(err)
	}

	total := res.Hits.TotalHits.Value
	var typ config.Result
	var da config.Data
	if total > 0 {
		num := 0
		for range res.Each(reflect.TypeOf(typ)) {
			num++
		}
		var primaryKey = make([]string, num)
		for i, item := range res.Each(reflect.TypeOf(typ)) { //从搜索结果中取数据的方法
			t := item.(config.Result)
			primaryKey[i] = t.PrimaryKey
		}
		da.Records = primaryKey
	} else {
		da.Records = []string{}
	}

	//组装返回结果
	da.Total = res.Hits.TotalHits.Value
	da.Page = int64(page)
	da.Size = int64(size)
	return da
}

//删除
func delete() {
	res, err := client.Delete().Index("megacorp").
		Id("1").
		Do(context.Background())
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("delete result %s\n", res.Result)
}
