/*
 * @Descripttion:
 * @version:
 * @Date: 2023-04-29 22:30:30
 * @LastEditTime: 2023-07-03 20:58:10
 */
package data

import (
	"context"
	"gpt-meeting-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMongo, NewRoleTemplateRepo, NewMeetingTemplateRepo, NewMeetingRepo, NewDifyRepo)

// Data .
type Data struct {
	mongodb *mongo.Database
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, mongodb *mongo.Database) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		mongodb: mongodb,
	}, cleanup, nil
}

func NewMongo(c *conf.Data) *mongo.Database {
	var clientOpts = options.Client().SetAuth(
		options.Credential{
			// AuthMechanism: "SCRAM-SHA-1",
			Username: c.Mongo.Username,
			Password: c.Mongo.Password,
		}).
		SetConnectTimeout(c.Mongo.ConnectTimeout.AsDuration()).
		SetHosts([]string{c.Mongo.Host}).
		SetMaxPoolSize(uint64(c.Mongo.MaxPoolSize)).
		SetMinPoolSize(uint64(c.Mongo.MinPoolSize)).
		SetMonitor(
			&event.CommandMonitor{
				// Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
				// 	log.Infof(cse.Command.String())
				// },
				Succeeded: func(ctx context.Context, cse *event.CommandSucceededEvent) {
					log.Infof(cse.Reply.String())
				},
				Failed: func(ctx context.Context, cfe *event.CommandFailedEvent) {
					log.Warnf(cfe.Failure)
				},
			},
		)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Errorf("failed to connect mongo: %v", err)
		panic("failed to connect mongo")
	}
	return client.Database(c.Mongo.Dbname)
}
