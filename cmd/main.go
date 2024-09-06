package main

import (
	"fmt"
	"log"
	"mainService/config"
	pb "mainService/genproto/doccs"
	"mainService/pkg/logger"
	"mainService/service"
	"mainService/storage/mongodb"
	"net"

	"google.golang.org/grpc"
)

func main() {
	log.Println("server is started")
	listener, err := net.Listen("tcp", config.Load().GOOGLE_DOCS)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	mongoDB, err := mongodb.ConnectMongoDb()
	if err != nil {
		log.Fatal(err)
	}

	logs := logger.NewLogger()
	mongodbRepoDocument := mongodb.NewDocumentRepository(mongoDB)
	mongodbRepoVersion := mongodb.NewDocumentVersionRepository(mongoDB)

	mongodbService := service.NewService(logs, mongodbRepoDocument, mongodbRepoVersion)

	server := grpc.NewServer()
	pb.RegisterDocsServiceServer(server, mongodbService)

	fmt.Printf("Server is listening on port %s\n", config.Load().GOOGLE_DOCS)
	if err = server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
