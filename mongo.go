package hs

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func ConnectMongo(ctx context.Context, conf MongoConf) (*mongo.Database, error) {
	option := options.Client().ApplyURI(conf.URI)
	if conf.MaxPoolSize > 0 {
		option.SetMaxPoolSize(conf.MaxPoolSize)
	}
	if conf.MinPoolSize > 0 {
		option.SetMinPoolSize(conf.MinPoolSize)
	}
	if len(conf.AppName) > 0 {
		option.SetAppName(conf.AppName)
	}

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		if client != nil {
			_ = client.Disconnect(ctx)
		}
		return nil, err
	}
	return client.Database(conf.Database), nil
}

func IsDuplicateError(err error) bool {
	e, ok := err.(mongo.WriteException)
	if !ok {
		return false
	}
	if e.WriteConcernError == nil && len(e.WriteErrors) == 1 && e.WriteErrors[0].Code == 11000 {
		return true
	}
	return false
}

func SaveInt64(ctx context.Context, coll *mongo.Collection, key string, value int64) error {
	option := &options.UpdateOptions{}
	option.SetUpsert(true)

	_, err := coll.UpdateOne(ctx,
		bson.D{
			{"key", key},
		},
		bson.D{
			{"$set", bson.D{
				{"value", value},
			}},
			{"$currentDate", bson.D{
				{"lastModified", true},
			}},
		},
		option,
	)
	return err
}

func LoadInt64(ctx context.Context, coll *mongo.Collection, key string) (int64, error) {
	var state = struct {
		Key          string
		Value        int64
		LastModified time.Time
	}{}
	if err := coll.FindOne(ctx, bson.D{
		{"key", key},
	}).Decode(&state); err == mongo.ErrNoDocuments {
		//s.Sugar.Errorf("load Position error: %s",err)
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return state.Value, nil
}

func DeleteInt64(ctx context.Context, coll *mongo.Collection, key string) error {
	_, err := coll.DeleteOne(ctx, bson.D{
		{"key", key},
	})
	return err
}

func DeleteKey(ctx context.Context, coll *mongo.Collection, key string) error {
	_, err := coll.DeleteOne(ctx, bson.D{
		{"name", key},
	})
	return err
}

func SaveKey(ctx context.Context, coll *mongo.Collection, key string, value interface{}) error {
	option := &options.UpdateOptions{}
	option.SetUpsert(true)

	_, err := coll.UpdateOne(ctx,
		bson.D{
			{"name", key},
		},
		bson.D{
			{"$set", bson.D{
				{"value", value},
			}},
			{"$currentDate", bson.D{
				{"lastModified", true},
			}},
		},
		option,
	)
	return err
}

func LoadKey(ctx context.Context, coll *mongo.Collection, key string, value interface{}) error {
	raw, err := coll.FindOne(ctx, bson.D{
		{"name", key},
	}).DecodeBytes()
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return err
	}
	doc, err := raw.LookupErr("value")
	if err != nil {
		return err
	}
	return doc.Unmarshal(value)
}
