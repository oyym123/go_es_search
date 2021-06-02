## 简介

- golang search 服务
- 主要是PHP 发送数据到rabbitmq
- 然后go 多线程监听rabbitmq获取数据 插入数据到ES
- PHP可以通过go 提供的Http服务进行ES搜索

## 安装

- 更改 config.dev.json中的相关db配置与项目路径配置
- 执行 go build -o build/EsSearch
- 执行 build/EsSearch


## 注意

目前go服务只会返回主键，可以通过主键再次搜索数据库，查出相应信息进行返回

## 使用

例如使用PHP客户端插入数据，以及搜索信息 

```php

<?php
/**
 * Created by PhpStorm.
 * Date: 2020/5/23
 * Time: 17:15
 */

class EsActionService extends ServiceAbstract
{
    public $config = [];                //整体配置文件 在 config/essearch.php
    public $type = 1;                   //类型 暂无作用默认=1 【预留】
    public $limit = 20;                 //每页数量
    public $page = 1;                   //分页
    public $singlePushNum = 200;        //单次推送封装数量
    public $orderField = 'primaryKey';  //排序字段
    public $orderType = 1;              //排序 1=升序 0=降序
    public $condition = [];             //搜索条件
    public $test = 0;                   //1=需要测试

    /**
     * EsSearch constructor.
     * @param string $confDefault
     */
    function __construct($confDefault)
    {
        $config = [];
        //直接引用 配置文件中的$config变量
        include APPPATH . "config/essearch.php";
        $this->config = $config['essearch'][$confDefault];
    }


    /**
     * 监听队列
     */
    public function listenEsMq()
    {
        $path = $this->config['url'] . '/MqConn';
        return $this->getRes($path, $this->config['rabbitmq']);
    }

    /**
     * 添加索引
     */
    public function addIndex()
    {
        $path = $this->config['url'] . '/CreateMapping';
        $data = [
            'index' => $this->config['index'],
            'mapping' => '{}'
        ];
        print_r(curl_request($path, json_encode($data)));
    }

    /**
     * 添加映射
     */
    public function addMapping()
    {
        $path = $this->config['url'] . '/CreateMapping';
        foreach ($this->config['addMapping'] as $value) {
            if ($value['status'] == 1) {
                $data = [
                    'index' => $this->config['index'],
                    'mapping' => $value['mapping'],
                ];
                print_r(curl_post($path, json_encode($data)));
            }
        }
    }

    /*
     * 搜索数据
     * @param array $search 搜索配置数组 【注意所有传值都应该是字符串，int，float都会被转成字符串】
     * condition:
     * [
            [
                'field' => 'titleCn',
                'search' => 'like',
                'value' => '11'
            ], [
                'field' => 'newPrice',
                'search' => 'between',
                'value_l' => '0',
                'value_r' => '1111'
            ],
            [
                'field' => 'categoryId',
                'search' => '=',
                'value' => '1765',
            ],
            [
                'path' => 'warehouse_price.stock', //json串子搜索
                'field' => 'warehouse_price',
                'search' => 'json_search',         //嵌套搜索
                'path_search' => 'between',       //子搜索类型  支持以下所有搜索类型
                'value_l' => '1',
                'value_r' => '100',
            ]
       ];
     支持的搜索类型 search:  = , > , >= , < , <=, != , like , not_like , in , not_in, between_left, between_right , between,json_search
     between_left : 左开区间   <= value <
     between_right :右开区间   < value <=
     between: 全开区间  <= value <=
     json_search : 嵌套搜索    [{"warehouse_code":"72","stock":405,"head_price":"2.95","tail_price":"58.37","total_price":"63.48","duty_cost":"2.16","extra_cost":"0.00"}]
     */
    public function search()
    {
        $path = $this->config['url'] . '/SizeSearch';
        $data = $tmp = [];

        //字符串转换
        foreach ($this->condition as $item) {
            foreach ($item as $k => $value) {
                if (is_array($value)) {
                    $value = json_encode($value);
                }
                $tmp[$k] = (string)$value;
            }
            $data[] = $tmp;
        }

        $search = [
            'name' => (string)$this->config['index'], (string)
            'condition' => $data,
            'type' => $this->type,
            'limit' => $this->limit > 10000 ? 10000 : intval($this->limit),
            'page' => intval($this->page),
            'order_field' => $this->orderField,
            'order_type' => intval($this->orderType),
        ];
        if ($this->test) {
            print_r($search);
            exit;
        }
        return $this->getRes($path, $search);
    }

    /**
     * 搜索字段校验
     * @param $searchList
     * @param $request
     * @return array
     */
    public function checkSearch($searchList, $request)
    {
        $searchArr = [];
        foreach ($request as $key => $value) {
            foreach ($searchList as $search) {
                //当填写了别名，则用别名匹配，否则用原本字段值匹配
                if ((isset($search['field_alias']) && $key == $search['field_alias']) || (!isset($search['field_alias']) && $key == $search['field'])) {
                    if ($value !== '') {
                        //主键转换
                        if ($key == $this->config['primaryKey']) {
                            $search['field'] = 'primaryKey';
                            if (!is_array($value)) {
                                $value = explode(',', $value);
                            }
                        }

                        if ($search['search'] == 'json_search') { //表示嵌套搜索
                            $searchArr[] = [
                                'field' => $search['field'],
                                'search' => $search['search'],
                                'value' => $value,
                                'path' => $search['path'],
                                'path_search' => $search['path_search'],
                            ];
                        } else {
                            if (strpos($search['search'], 'between') !== false) {
                                $v = explode(',', $value);
                                if (isset($v[0]) && isset($v[1])) {
                                    //将日期格式化到ES接受的格式
                                    if (isset($search['format']) && $search['format'] == 'time') {
                                        //适配前端只传日期
                                        if (mb_strlen($v[0]) < 11 && date('Y-m-d', strtotime($v[0])) == $v[0]) {
                                            $v[0] = $v[0] . ' 00:00:00';
                                        }

                                        if (mb_strlen($v[1]) < 11 && date('Y-m-d', strtotime($v[1])) == $v[1]) {
                                            $v[1] = $v[1] . ' 23:59:59';
                                        }
                                    }
                                    $searchArr[] = [
                                        'field' => $search['field'],
                                        'search' => $search['search'],
                                        'value_l' => $v[0],
                                        'value_r' => $v[1],
                                    ];
                                }
                            } elseif (strpos($search['search'], 'in') !== false) {
                                //当有逗号时，则转换为数组，如果已经是数组则不转换
                                if (!is_array($value)) {
                                    $value = explode(',', $value);
                                }

                                $searchArr[] = [
                                    'field' => $search['field'],
                                    'search' => $search['search'],
                                    'value' => $value,
                                ];
                            } else {
                                $searchArr[] = [
                                    'field' => $search['field'],
                                    'search' => $search['search'],
                                    'value' => $value,
                                ];
                            }
                        }
                    }
                }
            }
        }
        return $searchArr;
    }

    /**
     * 创建或更新数据
     * @param array[][] $data 二维数组
     * @return array
     */
    public function createOrUpdate($data)
    {
        $config = $this->config;
        $type = 1;
        $rabbitmq = load_library('rabbitmq', $config['mqType']);
        $arr = $mapping = [];
        $keys = array_unique(array_keys($data[0]));
        $addMapping = [];
        if (isset($config['addMapping']) && !empty($config['addMapping'])) {
            foreach ($config['addMapping'] as $k1 => $value) {
                $re = json_decode($value['mapping'], true)['properties'];
                foreach ($re as $k => $item) {
                    $addMapping[$k] = $item;
                }
            }
        }
        //获取所有的映射字段
        $mapKeys = array_keys($addMapping);

        $nestedField = [];
        //查询出所有的映射字段，以及是否是嵌套字段
        foreach ($addMapping as $key => $value) {
            if ($value['type'] == 'nested') {
                $nestedField[] = $key;
            }
        }

        $flag = 0;
        foreach ($keys as $str) {
            if (strpos($str, '_')) {
                $uncamelizedWords = '_' . str_replace('_', " ", strtolower($str));
                $mapping[$str] = ltrim(str_replace(" ", "", ucwords($uncamelizedWords)), '_');
                if (!in_array($mapping[$str], $mapKeys)) {
                    return [-1, '存在没有映射的字段' . $mapping[$str] . '，请添加映射后再推送'];
                }
            } else {
                $mapping[$str] = $str;
            }
            if (strpos($str, $config['primaryKey']) !== false) {
                $mapping[$str] = 'primaryKey';
                $flag = 1;
            }
        }

        if (empty($flag)) {
            return [-1, '必须填写主键字段' . $config['primaryKey']];
        }

        //将字段名称转成单驼峰
        foreach ($data as $k => $re) {
            $t = [];
            foreach ($re as $key => $value) {
                if (in_array($key, array_keys($mapping))) {
                    if (in_array($mapping[$key], $nestedField)) {  //嵌套字段，进行修改
                        if (($value !== '' && $value !== null) || $value === 0) {
                            if ($value === 0) {
                                $value = "0";
                            }
                            $jsonStr = json_decode((string)$value);
                            if (!empty($jsonStr)) {
                                $t[$mapping[$key] . 'NestedEs'] = [$mapping[$key] => $jsonStr];
                            } else {
                                return [-1, $mapping[$key] . ' 字段json解析错误！' . $jsonStr];
                            }
                        }
                    } else {
                        if (($value !== '' && $value !== null) || $value === 0) {
                            if ($value === 0) {
                                $value = "0";
                            }
                            //加上对应的类型
                            $t[$mapping[$key]] = $value;
                        }
                    }
                }
            }
            $arr[] = $t;
        }

        if ($this->test == 1) {
            return $arr;
        }

        //分批推送到队列
        $allNum = count($arr);
        foreach ($arr as $key => $value) {
            $saveData[] = $value;
            //当循环到不够一次封装的量时，则单条推送出去
            if ($allNum - $key <= $this->singlePushNum + 1) {
                $newArr = [
                    'type' => $type,
                    'name' => $config['index'],
                    'content' => $saveData
                ];

                $rabbitmq->sendMsg(json_encode($newArr), $config['rabbitmq']['ex_name'], $config['rabbitmq']['qu_name']);
                $saveData = [];
                continue;
            }
            if ($key % $this->singlePushNum == 0) {
                $newArr = [
                    'type' => $type,
                    'name' => $config['index'],
                    'content' => $saveData
                ];
                $rabbitmq->sendMsg(json_encode($newArr), $config['rabbitmq']['ex_name'], $config['rabbitmq']['qu_name']);
                $saveData = [];
            }
        }
        return [1, '推送成功!'];
    }


    public function getRes($path, $data)
    {
        return curl_request($path, json_encode($data));
    }
}

```

essearch.php

```
//产品模块ES映射
$product = [
    'url' => $urlEnv,
    'index' => $indexEnv,
    'mqType' => $env,  //mq使用的配置 默认 []
    'rabbitmq' => [
        'ex_name' => 'GO_ES_SEARCH',
        'qu_name' => 'dcm_product_es_search',
        're_key' => 'dcm_product_es_search',
        'ex_type' => 'direct',
        'num' => 2,   //线程数量 建议最大不超过 20
    ], //初始化映射
    'primaryKey' => 'sku', //将sku字段作为主键
    'addMapping' => [  //后续添加的映射
        1 => [
            'mapping' => '{   
                "properties":{
                  "avgPrice": {
                    "type": "float"
                  }
			    }
             }',
            'status' => 1,  // 1=需要执行  0=已经执行完毕
        ],
        2 => [
            'mapping' => '{
                "properties":{
                  "shipCost": {
                  "type": "float"
                  }
			    }
             }',
            'status' => 1,  // 1=需要执行  0=已经执行完毕
        ],
        3 => [
            'mapping' => '{
                "properties":{
                  "fee": {
                  "type": "float"
                  },
                  "isFba": {
                  "type": "integer"
                  },
                  "distributorPlatforms": {
                  "type": "text"
                  },
                  "endTime": {
                  "type": "date",
                  "format" : "yyyy-MM-dd HH:mm:ss"
                  },
                  "updatedAt": {
                  "type": "date",
                  "format" : "yyyy-MM-dd HH:mm:ss"
                  },
                  "weightOutStorage": {
                  "type": "float"
                  },
                  "netWeight": {
                  "type": "float"
                  },
                  "grossWight": {
                  "type": "float"
                  },
                  "resourceType": {
                  "type": "float"
                  },
                  "isOversea": {
                  "type": "integer"
                  },
                  "warehousePrice": {
                  "type": "text"
                  },
                  "purchasePrice": {
                  "type": "float"
                  },
                  "attrValueIds": {
                  "type": "float"
                  },
                  "createdAt": {
                  "type": "date",
                  "format" : "yyyy-MM-dd HH:mm:ss"
                  },
                  "sku": {
                  "type": "keyword",
                  "fields": {
                    "keyword": {
                      "ignore_above": 256,
                      "type": "keyword"
                    }
                 }
                  },
                  "status": {
                  "type": "integer"
                  },
                  "resourceStatus": {
                  "type": "integer"
                  },
                  "productStatus": {
                  "type": "integer"
                  }
			    }
             }',
            'status' => 1,  // 1=需要执行  0=已经执行完毕
        ],
        4 => [
            'mapping' => '{
                "properties":{
                  "dcmLevel": {
                  "type": "integer"
                  },
                  "attrIds": {
                      "type": "nested",
                       "properties" : {
                          "id" : {
                            "type" : "integer"
                            }
                          }
                    }
                  }
             }',
            'status' => 1,  // 1=需要执行  0=已经执行完毕
        ],
        5 => [
            'mapping' => '{ 
            "properties": {
              "titleEn": {
              "type": "text",
              "fields": {
                "keyword": {
                  "ignore_above": 500,
                  "type": "keyword"
                }
              }
            },
            "titleCn": {
              "type": "text",
              "fields": {
                "keyword": {
                  "ignore_above": 256,
                  "type": "keyword"
                }
              }
            },
            "categoryStr": {
              "type": "text",
              "fields": {
                "keyword": {
                  "ignore_above": 256,
                  "type": "keyword"
                }
              }
            },
            "distributorNumbers": {
              "type": "text"
            },
            "infringementRemarks": {
              "type": "keyword",
              "fields": {
                "keyword": {
                  "ignore_above": 2000,
                  "type": "keyword"
                }
              }
            },
            "newPrice": {
              "type": "float"
            },
            "categoryId": {
              "type": "text",
              "fields": {
                "keyword": {
                  "ignore_above": 256,
                  "type": "keyword"
                }
              }
            },
            "primaryKey": {
              "type": "keyword",
              "fields": {
                "keyword": {
                  "ignore_above": 256,
                  "type": "keyword"
                }
              }
            }
          }
      }',
            'status' => 1,  // 1=需要执行  0=已经执行完毕
        ],
    ]
];

$config["essearch"]["product"] = $product;


```

test.php
```
        $es = new EsActionService();
        //搜索项
        $condition = [
                 [
                     'field' => 'sku',
                     'search' => 'in',
                 ],
                 [
                     'field' => 'titleEn',
                     'search' => 'like',
                 ],
                 [
                     'field' => 'titleCn',
                     'search' => 'like',
                 ],
                 [
                     'field' => 'categoryStr',
                     'search' => 'like',
                 ],
                 [
                     'field' => 'updatedAt',
                     'search' => 'between',
                     'format' => 'time'
                 ],
                 [
                     'field' => 'weightOutStorage',
                     'search' => 'between',
                 ],
                 [
                     'field' => 'netWeight',
                     'search' => 'between',
                 ],
                 [
                     'field' => 'grossWight',
                     'search' => 'between',
                 ],
                 [
                     'field' => 'resourceType',
                     'search' => '=',
                 ],
                 [
                     'field' => 'dcmLevel',
                     'search' => 'in',
                 ],
                 [
                     'field' => 'categoryId',
                     'search' => 'in',
                 ],
                 [
                     'field' => 'isOversea',
                     'search' => '=',
                 ],
                 [
                     'field' => 'status',
                     'search' => '=',
                 ],
                 [
                     'field' => 'distributorNumbers',
                     'search' => 'like',
                 ],
                 [
                     'field_alias' => 'notDistributorNumbers',
                     'field' => 'distributorNumbers',
                     'search' => 'not_like',
                 ],
                 [
                     'field' => 'warehousePrice',
                     'search' => 'like',
                 ],
                 [
                     'field_alias' => 'notWarehousePrice',
                     'field' => 'warehousePrice',
                     'search' => 'not_like',
                 ],
                 [
                     'field_alias' => 'availableWarehouse',
                     'field' => 'warehousePrice',
                     'search' => 'like',
                 ],
                 [
                     'field' => 'purchasePrice',
                     'search' => 'between',
                 ],
                 [
                     'field_alias' => 'purchasePrice1',
                     'field' => 'purchasePrice',
                     'search' => 'should_between',
                 ],
                 [
                     'field_alias' => 'purchasePrice2',
                     'field' => 'purchasePrice',
                     'search' => 'should_between',
                 ],
                 [
                     'field_alias' => 'purchasePrice3',
                     'field' => 'purchasePrice',
                     'search' => 'should_lt',
                 ],
                 [
                     'field_alias' => 'purchasePrice4',
                     'field' => 'purchasePrice',
                     'search' => 'should_gt',
                 ],
                 [
                     'field' => 'distributorPlatforms',
                     'search' => 'like',
                 ],
                 [
                     'field' => 'notDistributorPlatforms',
                     'search' => 'not_like',
                 ],
                 [
                     'field' => 'createdAt',
                     'search' => 'between',
                     'format' => 'time'
                 ],
                 [
                     'field' => 'endTime',
                     'search' => 'between_left',
                     'format' => 'time'   //格式化时间 2020-11-01 22:01:00 兼容前端只传日期ES搜索不到,后端自动加上时分秒
                 ],
                 [
                     'field' => 'isFba',
                     'search' => '=',
                 ],
                 [
                     'field' => 'infringementRemarks',
                     'search' => '=',
                 ],
                 [
                     'path' => 'attrIds.id', //json串子搜索
                     'field' => 'attrIds',
                     'search' => 'json_search',         //嵌套搜索
                     'path_search' => 'in',             //子搜索类型  支持以下所有搜索类型
                 ]
             ];

             $request = [
                'sku' => '1111'
             ];

             $es->condition = $es->checkSearch($condition, $request);
             $res = $es->search();
             print_r($res);
```
