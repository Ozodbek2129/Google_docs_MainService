package main

import (
	"log"
	"mainService/config"
	"mainService/storage/mongodb"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", config.Load().GOOGLE_DOCS)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	mongodb, err := mongodb.ConnectMongoDb()
	if err != nil {
		log.Fatal(err)
	}

	
}
