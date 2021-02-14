package main

import (
	"context"
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
	header, err := resp.Header()
	if err != nil {
		log.Printf("when read header, got: %v", err)
	}
	if a, ok := header["application"]; ok {
		for _, v := range a {
			log.Printf("host: %v\n", v)
		}
	} else {
		log.Println("wait application field but doesn't get")
	}

	for {
		msg, err := resp.Recv()
		if err != nil {
			log.Fatalf("when read message, got: %v", err)
		}
		log.Printf(" - \n%v\n", msg)
	}
}
