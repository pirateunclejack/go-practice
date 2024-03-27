package main

import (
	"context"
	"log"
	"time"

	pb "github.com/pirateunclejack/go-practice/grpc-demo-yt/proto"
)

func callSayHello(client pb.GreetServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	res, err := client.SayHello(ctx, &pb.NoParam{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("response: %s", res.Message)
}
