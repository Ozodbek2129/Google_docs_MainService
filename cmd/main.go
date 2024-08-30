package main

import (
	"log"
	"mainService/config"
	"mainService/pkg/logger"
	"mainService/service"
	"mainService/storage/mongodb"
	pb "mainService/genproto/docs"
	"google.golang.org/grpc"
	"net"
)

func main() {
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
	pb.RegisterDocsServiceServer(server,mongodbService)

	log.Printf("Server is listening on port %s\n", config.Load().GOOGLE_DOCS)
	if err = server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
