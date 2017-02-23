package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"

	"strings"

	"google.golang.org/grpc"

	"golang.org/x/net/context"
	"golang.org/x/net/websocket"

	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
	grpchello "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
)

type exampleHTTPHandler struct{}

func (h *exampleHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "example http response")
}

func serveHTTP(l net.Listener) {
	s := &http.Server{
		Handler: &exampleHTTPHandler{},
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}

func EchoServer(ws *websocket.Conn) {
	if _, err := io.Copy(ws, ws); err != nil {
		panic(err)
	}
}

func serveWS(l net.Listener) {
	s := &http.Server{
		Handler: websocket.Handler(EchoServer),
	}
	if err := s.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}

type ExampleRPCRcvr struct{}

func (r *ExampleRPCRcvr) Cube(i int, j *int) error {
	*j = i * i
	return nil
}

func serveRPC(l net.Listener) {
	s := rpc.NewServer()
	if err := s.Register(&ExampleRPCRcvr{}); err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			if err != cmux.ErrListenerClosed {
				panic(err)
			}
			return
		}
		go s.ServeConn(conn)
	}
}

type grpcServer struct{}

func (s *grpcServer) SayHello(ctx context.Context, in *grpchello.HelloRequest) (
	*grpchello.HelloReply, error) {

	metadata.FromContext(ctx)

	return &grpchello.HelloReply{Message: "Hello " + in.Name + " from cmux"}, nil
}

func (s *grpcServer) Send(stream grpchello.Greeter_SendServer) error {

	for {
		select {

		case <-stream.Context().Done():
			//fmt.Println("client connection closed")
			break

		default:
			req, err := stream.Recv()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Sprintf("req %s", req.Name)
			break

		}
	}

}

func serveGRPC(l net.Listener) {
	grpcs := grpc.NewServer()
	grpchello.RegisterGreeterServer(grpcs, &grpcServer{})
	if err := grpcs.Serve(l); err != cmux.ErrListenerClosed {
		panic(err)
	}
}

func Example() {
	l, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Panic(err)
	}

	m := cmux.New(l)

	// We first match the connection against HTTP2 fields. If matched, the
	// connection will be sent through the "grpcl" listener.
	//grpcl := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))

	grpcl := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	//Otherwise, we match it againts a websocket upgrade request.
	wsl := m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))

	// Otherwise, we match it againts HTTP1 methods. If matched,
	// it is sent through the "httpl" listener.
	httpl := m.Match(cmux.HTTP1Fast())
	// If not matched by HTTP, we assume it is an RPC connection.
	rpcl := m.Match(cmux.Any())

	// Then we used the muxed listeners.
	go serveGRPC(grpcl)
	go serveWS(wsl)
	go serveHTTP(httpl)
	go serveRPC(rpcl)

	p := NewPprof()
	go p.Start()

	if err := m.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		panic(err)
	}
}

type pprof struct {
}

func YourHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pprof handler called")
}

func NewPprof() *pprof {
	return &pprof{}
}

func (p *pprof) Start() error {
	fmt.Println("Starting pprof")
	r := mux.NewRouter()
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	r.HandleFunc("/hello", YourHandler)
	http.ListenAndServe("127.0.0.1:8000", r)
	return nil
}

func main() {
	Example()
}
