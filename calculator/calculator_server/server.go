package main

import (
	"context"
	"github.com/anthonysyk/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
)

type server struct{}

func (s server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {

	log.Printf("Request received: %v", req)
	sum := req.Number_1 + req.Number_2
	res := &calculatorpb.SumResponse{
		Result: sum,
	}

	return res, nil
}

func (s server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	log.Printf("RPC Call PrimeNumberDecomposition invoked for number: %v", req.GetNumber())
	primes := PrimeFactors(int(req.GetNumber()))

	for _, p := range primes {
		log.Printf("Sending prime factor: %v", p)
		err := stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{Number: int64(p)})
		if err != nil {
			log.Fatalf("error on sending RPC PrimeNumberDecomposition response: %v", err)
		}
	}

	return nil
}

func (s server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	var values []float64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error on sending RPC Average response: %v", err)
		}

		log.Printf("Received number: %v", req.GetNumber())
		values = append(values, req.GetNumber())
	}

	res := Average(values)

	err := stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
		Result: res,
	})

	if err != nil {
		log.Fatalf("error on sending RPC Average response: %v", err)
	}

	return nil
}

func (s server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	max := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error on receiving request: %v", err)
			return err
		}
		log.Printf("Received: %v", req.GetNumber())

		if req.GetNumber() > max {
			max = req.GetNumber()
			log.Printf("Maximum is: %v", max)
			err = stream.Send(&calculatorpb.FindMaximumResponse{
				Result: max,
			})
		}
		if err != nil {
			log.Fatalf("error on sending mesasge: %v", err)
			return err
		}
	}

	return nil
}

func (s server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Received number is negative: %v", number)
	}

	return &calculatorpb.SquareRootResponse{
		Result: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, server{})
	reflection.Register(s)

	log.Println("Server calculator listening ...")

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

// Get all prime factors of a given number n
func PrimeFactors(n int) (pfs []int) {
	// Get the number of 2s that divide n
	for n%2 == 0 {
		pfs = append(pfs, 2)
		n = n / 2
	}

	// n must be odd at this point. so we can skip one element
	// (note i = i + 2)
	for i := 3; i*i <= n; i = i + 2 {
		// while i divides n, append i and divide n
		for n%i == 0 {
			pfs = append(pfs, i)
			n = n / i
		}
	}

	// This condition is to handle the case when n is a prime number
	// greater than 2
	if n > 2 {
		pfs = append(pfs, n)
	}

	return
}

func Average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}
