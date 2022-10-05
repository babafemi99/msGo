package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"logger-srv/cmd/data"
	"logger-srv/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	//Write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err = l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		return nil, err
	}

	//return response
	return &logs.LogResponse{Result: "logged"}, nil
}

func (c *Config) grpcListen() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen for grpc: %v", err)
	}
	defer listen.Close()
	s := grpc.NewServer()
	logs.RegisterLogServiceServer(s, &LogServer{Models: c.Models})
	log.Printf("grpc server started on port: %s", grpcPort)

	err = s.Serve(listen)
	if err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
