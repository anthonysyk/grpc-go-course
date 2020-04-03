package main

import (
	"context"
	"fmt"
	"github.com/anthonysyk/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	cc, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v ", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doUnary(c)
	// doServerStreaming(c)
	// doClientStreaming(c)
	// doBiDiStreaming(c)
	doErrorUnary(c)
}

func doUnary(client calculatorpb.CalculatorServiceClient) {
	res, err := client.Sum(context.Background(), &calculatorpb.SumRequest{
		Number_1: 3,
		Number_2: 10,
	})

	if err != nil {
		log.Fatalf("failed RPC Call Sum(): %v", err)
		return
	}

	log.Printf("Sum result: %d", res.Result)
}

func doServerStreaming(client calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 120,
	}

	stream, err := client.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error on RPC call PrimeNumberDecomposition: %v", err)
	}

	var acc []int64
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}

		log.Printf("PrimeDecomposition number received: %v", res.GetNumber())
		acc = append(acc, res.GetNumber())
		time.Sleep(1000 * time.Millisecond)
	}

	log.Printf("PrimeDecomposition numbers: %v", acc)
}

func doClientStreaming(client calculatorpb.CalculatorServiceClient) {
	stream, err := client.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error on RPC call ComputeAverage: %v", err)
	}

	requests := []*calculatorpb.ComputeAverageRequest{
		{Number: 1},
		{Number: 2},
		{Number: 3},
		{Number: 4},
	}

	for _, req := range requests {
		err := stream.Send(req)
		if err != nil {
			log.Fatalf("error while sending stream: %v", err)
		}
		log.Printf("Average number sent: %v", req.GetNumber())
		time.Sleep(1000 * time.Millisecond)
	}

	err = stream.CloseSend()
	if err != nil {
		log.Fatalf("error while closing stream")
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response: %v", err)
	}

	log.Printf("Average of %v is: %v", requests, res.GetResult())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("error on RPC call FindMaximum: %v", err)
	}

	requests := []*calculatorpb.FindMaximumRequest{{Number: 1}, {Number: 5}, {Number: 3}, {Number: 6}, {Number: 2}, {Number: 20}}

	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("error while sending request: %v", err)
			}
			log.Printf("Sending: %v", req.GetNumber())
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error on receiving response: %v", err)
			}

			log.Printf("Current maximum is: %v", res.GetResult())
		}
	}()

	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	squareRoot(10, c)
	squareRoot(-2, c)
}

func squareRoot(n int32, c calculatorpb.CalculatorServiceClient) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: n,
	})
	log.Printf("Sent number: %v", n)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			fmt.Printf("Error code: %v \n", s.Code())
			fmt.Printf("Error message: %v \n", s.Message())
			return
		} else {
			log.Fatalf("An error occured with RPC call SquareRoot: %v", err)
			return
		}
	}
	fmt.Printf("Square Root for number %v is %v \n", n, res.GetResult())
}
