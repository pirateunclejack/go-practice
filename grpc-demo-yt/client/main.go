package main

import (
	"log"

	pb "github.com/pirateunclejack/go-practice/grpc-demo-yt/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
    port = ":8888"
)

func main() {
    conn, err := grpc.Dial(
        "localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("failed to connect grpc server: %v", err)
    }
    defer conn.Close()

    client := pb.NewGreetServiceClient(conn)

    names := &pb.NamesList{
        Names: []string{
            "name1",
            "name2",
            "name3",
        },
    }

    // callSayHello(client)
    // callSayHelloServerStream(client, names)
    // callSayHelloClientStream(client, names)
    callHelloBidirectionalStream(client, names)
}