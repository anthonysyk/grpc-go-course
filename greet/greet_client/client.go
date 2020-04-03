package main

import (
	"context"
	"fmt"
	"github.com/anthonysyk/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {

	opts := []grpc.DialOption{}
	tls := false
	if tls {
		creds, sslErr := credentials.NewClientTLSFromFile("ssl/ca.crt", "")
		if sslErr != nil {
			log.Fatalf("Failed to load certificates= %v", sslErr)
			return
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	cc, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	fmt.Println("Created client: %f", c)

	doUnary(c)
	// doServerStreaming(c)
	// doClientStreaming(c)
	// doBiDiStreaming(c)
	// doUnaryWithDeadline(c, 1*time.Second)
	// doUnaryWithDeadline(c, 5*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Anthony",
			LastName:  "SSI YAN KAI",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error on Greet RPC call: %v", err)
	}

	log.Printf("RPC Call Greet: %v", res.GetResult())
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{
		FirstName: "Anthony",
		LastName:  "SSI YAN KAI",
	}

	stream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error on GreetManyTimes RPC call")
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	requests := []*greetpb.LongGreetRequest{
		{FirstName: "Anthony"},
		{FirstName: "Alex"},
		{FirstName: "Edouard"},
		{FirstName: "Lucy"},
		{FirstName: "Piper"},
		{FirstName: "Sofia"},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while creating Client Streaming RPC: %v", err)
	}

	for _, req := range requests {
		log.Printf("Sending Request: %v", req)
		err := stream.Send(req)
		if err != nil {
			log.Fatalf("error while sending LongGreetRequest: %v", err)
		}
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while closing streaming: %v", err)
	}

	fmt.Println(res)
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	requests := []*greetpb.GreetEveryoneRequest{
		{Greeting: &greetpb.Greeting{FirstName: "Anthony", LastName: "Test"}},
		{Greeting: &greetpb.Greeting{FirstName: "Alex", LastName: "Test"}},
		{Greeting: &greetpb.Greeting{FirstName: "Edouard", LastName: "Test"}},
		{Greeting: &greetpb.Greeting{FirstName: "Lucy", LastName: "Test"}},
		{Greeting: &greetpb.Greeting{FirstName: "Piper", LastName: "Test"}},
		{Greeting: &greetpb.Greeting{FirstName: "Sofia", LastName: "Test"}},
	}

	// init RPC call
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error while creating BiDi Streaming RPC: %v", err)
	}

	waitc := make(chan struct{})

	// send messages to server
	go func() {
		for _, req := range requests {
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("error while sending request: %v", err)
			}
			log.Printf("Sending: %v", req.GetGreeting().GetFirstName())
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// receive messages from server
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error while receiving respons: %v", err)
			}
			log.Printf("Receiving: %v", res.GetResult())
		}
		close(waitc)
	}()
	// block until end of stream
	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Printf("Sending RPC UnaryWithDeadline: with timeout of %v \n", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Anthony",
			LastName:  "SSI YAN KAI",
		},
	}

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			fmt.Printf("Error: %v - %v \n", statusErr.Code(), statusErr.Message())
		} else {
			fmt.Printf("Error: %v \b", err)
		}
		return
	}

	fmt.Printf("Result: %v \n", res.GetResult())
}
