package config

const ProductMapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
			"properties":{
				"primaryKey":{
					"type":"keyword",
					"fields": {
						"keyword": {
							"ignore_above": 256,
							"type": "keyword"
						}
					}
				},
				"titleEn": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 500,
							"type": "keyword"
						}
					}
				},
		        "warehousePrice": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 2000,
							"type": "keyword"
						}
					}
				},
                "distributorNumbers": {
					"type": "text"
		 		},
               "newPrice": {
					"type": "float"
				}
			}
	}
}`

const ProductMappingAdd = `{
			"properties":{
				"fbaCountryInfo":{
					"type":"nested"
				}
			}
}`

const ProductMappingAdd2 = `{
			"properties":{
				"assistantImg":{
					"type":"nested"
				}
			}
}`
