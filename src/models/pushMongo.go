package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"push_service/src/config"
	"push_service/src/core"
	"time"
)

type PushMongoModel struct {
	ID         int64
	Title      string
	Content    string
	Options    string
	MsgType    int
	UserIds    string
	SenderId   int64
	SenderName string
	CreateTime string
}

func (PushMongoModel) Create(m PushMessageModel) {
	mongoConf := core.Config.MongoDB
	for _, conf := range mongoConf {
		client, err := mongo.NewClient(options.Client().ApplyURI(conf.ApplyURI))
		err = client.Connect(ctx)
		defer func() {
			if err = client.Disconnect(ctx); err != nil {
				panic(err)
			}
		}()

		collection_name := fmt.Sprintf(config.COLLECTION_PRFIX+"%s", m.SenderName)
		collection := client.Database(conf.Database).Collection(collection_name)
		ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
		var _, _ = collection.InsertOne(ctx, &m)
		//fmt.Println(res.InsertedID)
	}
}
