package main

import (
	"context"
	"fmt"
	"log"

	smgrpc "github.com/hrapovd1/final_project/pkg/smgrpc"
	"google.golang.org/grpc"
)

func main() {

	conn, _ := grpc.Dial("127.0.0.1:8080", grpc.WithInsecure())

	client := smgrpc.NewStatClient(conn)

	resp, err := client.GetAll(context.Background(),
		&smgrpc.Request{Sent: true})
	if err != nil {
		log.Fatalf("could not get answer: %v", err)
	}
	for {
		msg, err := resp.Recv()
		if err != nil {
			log.Fatalf("when read message, got: %v", err)
		}
		fmt.Printf(" - %s\n", msg)
	}
}
