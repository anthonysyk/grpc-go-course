package main

import (
	"context"
	"fmt"
	"github.com/anthonysyk/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"time"
)

type server struct{}

func (s server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("Greet RPC was invoked: %v \n", req)
	greet := fmt.Sprintf("Hello %s", req.Greeting.GetFirstName())
	res := &greetpb.GreetResponse{
		Result: greet,
	}
	return res, nil
}

func (s server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Printf("GreetManyTimes RPC was invoked: %v \n", req)

	for i := 0; i < 10; i++ {
		log.Printf("Hello %s, number: %d \n", req.GetFirstName(), i)
		res := &greetpb.GreetManyTimesResponse{
			Result: req.GetFirstName(),
		}
		_ = stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

func (s server) LongGreet(c greetpb.GreetService_LongGreetServer) error {
	log.Printf("LongGreet RPC was invoked")
	result := ""

	for {
		req, err := c.Recv()
		if err == io.EOF {
			return c.SendAndClose(&greetpb.LongGreetResponse{Result: result})
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
		}
		current := fmt.Sprintf("Hello %s ! ", req.GetFirstName())
		fmt.Printf("Received LongGreet: %s \n", req.GetFirstName())
		result += current
	}
}

func (s server) GreetEveryone(c greetpb.GreetService_GreetEveryoneServer) error {
	log.Printf("GreetEveryone RPC was invoked")

	for {
		req, err := c.Recv()
		result := "Hello " + req.GetGreeting().GetFirstName() + "! "
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
			return err
		}

		log.Printf("Received request: %v", req)
		err = c.Send(&greetpb.GreetEveryoneResponse{Result: result})
		if err != nil {
			log.Fatalf("error while sending response: %v", err)
			return err
		}
	}

	return nil
}

func (s server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("Client canceled request")
			return nil, status.Error(codes.Canceled, "Client canceled request")
		}
		time.Sleep(1 * time.Second)
	}

	res := &greetpb.GreetWithDeadlineResponse{
		Result: fmt.Sprintf("Hello %v!", req.GetGreeting().GetFirstName()),
	}

	return res, nil
}

func main() {
	fmt.Println("Hello world")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listend: %v", err)
	}

	opts := []grpc.ServerOption{}
	tls := false
	if tls {
		creds, sslErr := credentials.NewServerTLSFromFile("ssl/server.crt", "ssl/server.pem")
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(s, server{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
