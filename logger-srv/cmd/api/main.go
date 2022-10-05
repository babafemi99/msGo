package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"logger-srv/cmd/data"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	webPort  = "80"
	rpcPort  = "50051"
	mongoUrl = "mongodb://mongo:27017"
	grpcPort = "50001"
)

type Config struct {
	Models data.Models
}

var (
	client *mongo.Client
	err    error
)

func main() {
	// connect to mongo
	client, err = connectToMongo()
	if err != nil {
		log.Panicln(err)
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*15)
	defer cancelFunc()

	defer func() {
		err = client.Disconnect(ctx)
		if err != nil {
			panic(err)
		}
	}()
	c := Config{Models: data.New(client)}

	err = rpc.Register(new(RPCServer))
	go c.listenRPC()
	go c.grpcListen()

	//go app.Serve()
	log.Println("starting service on port:", webPort)
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: c.Routes(),
	}
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

func (c *Config) listenRPC() {
	log.Println("Listen RPC server on PORT", rpcPort)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))
	if err != nil {
		return
	}
	defer listen.Close()
	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	connect, err := mongo.Connect(context.TODO(),
		options.Client().ApplyURI("mongodb://admindb:secret@mongo:27017"))
	if err != nil {
		log.Println("Error connecting to mongo DB")
		return nil, err
	}
	log.Println("connected to mongo")
	return connect, nil
}
