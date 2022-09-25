package service

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type MongoConnector struct {
	Username string
	Password string
	addrs    []string
	current  string
}

func NewMongoBalancer(username, password string, addrs []string) *MongoConnector {
	return &MongoConnector{Username: username,
		Password: password,
		addrs:    addrs,
	}
}

func (m *MongoConnector) CheckHealth(addr string) bool {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	opt := options.Client()
	opt.Direct = toBoolPtr(true)
	opt.ApplyURI("mongodb://" + m.Username + ":" + m.Password + "@" + addr + "/admin")
	//client, err := mongo.NewClient(opt)
	_, err := mongo.Connect(ctx, opt)
	if err != nil {
		return false
	}
	return true
}

func (m *MongoConnector) ChooseBackEnd() string {
	//if lb.current != "" {
	//	return lb.current
	//}
	for _, upstream := range m.addrs {
		if isMasterSync(upstream, m.Username, m.Password) {
			m.current = upstream
			return upstream
		}
	}
	return ""
}

func isMasterSync(addr string, user string, pwd string) bool {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	opt := options.Client()
	opt.Direct = toBoolPtr(true)
	opt.ApplyURI("mongodb://" + user + ":" + pwd + "@" + addr + "/admin")
	//client, err := mongo.NewClient(opt)
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return false
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Nearest())
	if err != nil {
		return false
	}
	result := client.Database("admin").RunCommand(ctx, bson.M{"isMaster": 1})
	var rst bson.M
	result.Decode(&rst)

	return rst["ismaster"].(bool)
}

func toBoolPtr(b bool) *bool {
	return &b
}
