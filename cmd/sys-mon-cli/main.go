package main

import (
	"context"
	"log"
	"os"

	ps "github.com/hrapovd1/final_project/pkg/smgrpc"

	"google.golang.org/grpc"
)

func main() {

	conn, _ := grpc.Dial("127.0.0.1:8080", grpc.WithInsecure())

	client := ps.NewStatClient(conn)

	resp, err := client.GetAll(context.Background(),
		&ps.Request{Sent: true})

	if err != nil {
		log.Fatalf("could not get answer: %v", err)
	}
	log.Println("New password:", resp.Password)
}
