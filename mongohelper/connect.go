package mongohelper

import (
	"context"
	"github.com/xyths/hs/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo(ctx context.Context, conf config.MongoConf) (*mongo.Database, error) {
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
		_ = client.Disconnect(ctx)
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
