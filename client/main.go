package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	"google.golang.org/grpc/metadata"
)

const (
	address            = "localhost:50051"
	defaultName        = "world"
	numConnections int = 100
)

var (
	grpcConnections []*grpcConnection
)

type grpcConnection struct {
	connection *grpc.ClientConn
	ctx        context.Context
	tenantID   string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

func launchGRPCConnections() {
	var copts []grpc.DialOption
	copts = append(copts, grpc.WithInsecure())

	zoneID := "topicA"

	token := strings.Repeat("s", 1024*1024)

	for i := 0; i < numConnections; i++ {

		conn, _ := grpc.Dial(address, copts...)
		headers := metadata.MD{
			"tenantid":      []string{"tenantid"},
			"authorization": []string{token},
			"content-type":  []string{"application/grpc"},
		}

		ctx := metadata.NewContext(context.Background(), headers)
		cancelCtx, cancelFunc := context.WithCancel(ctx)
		grpcConnections = append(grpcConnections, &grpcConnection{conn, ctx, zoneID, cancelCtx, cancelFunc})

	}

}

func dialServer2() {

	defer func() {
		for _, conn := range grpcConnections {
			if conn.connection != nil {
				conn.connection.Close()
			}
		}
	}()

	name := strings.Repeat("s", 1000)
	var copts []grpc.DialOption
	copts = append(copts, grpc.WithInsecure())
	token := strings.Repeat("s", 1024*1024)

	for i := 0; i < numConnections; i++ {

		conn, _ := grpc.Dial(address, copts...)
		headers := metadata.MD{
			"tenantid":      []string{"tenantid"},
			"authorization": []string{token},
			"content-type":  []string{"application/grpc"},
		}

		ctx := metadata.NewContext(context.Background(), headers)

		c := pb.NewGreeterClient(conn)
		stream, _ := c.Send(ctx)

		fmt.Println(stream)

		for j := 0; j < 1000; j++ {
			stream.Send(&pb.HelloRequest{Name: name})
		}
	}
}

func dialServer() {

	launchGRPCConnections()

	defer func() {
		for _, conn := range grpcConnections {
			if conn.connection != nil {
				conn.connection.Close()
			}
		}
	}()

	name := strings.Repeat("s", 1000)

	for _, conn := range grpcConnections {
		if conn.connection != nil {

			c := pb.NewGreeterClient(conn.connection)
			stream, _ := c.Send(conn.ctx)

			for j := 0; j < 1000; j++ {
				stream.Send(&pb.HelloRequest{Name: name})

			}
		}
	}
}

func main() {
	fmt.Println("Test mem growth called")
	time.Sleep(time.Second * 2)
	dialServer()
}
