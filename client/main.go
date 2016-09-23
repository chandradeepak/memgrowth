package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func dialServer() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	name = strings.Repeat("s", 100000)
	token := strings.Repeat("s", 1024*1024)

	headers := metadata.MD{
		"tenantid":      []string{"zoneID"},
		"authorization": []string{token},
		"content-type":  []string{"application/grpc"},
	}
	ctx := metadata.NewContext(context.Background(), headers)

	_, err = c.SayHello(ctx, &pb.HelloRequest{Name: name})
	//if err != nil {
	//log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.Message)
}

func main() {
	fmt.Println("Test mem growth called")
	time.Sleep(time.Second * 2)
	for i := 0; i < 100000; i++ {
		go dialServer()
	}

	time.Sleep(time.Second * 120)
}
