package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/dev/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

func (se *server) Sum(ctx context.Context, in *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	log.Printf("Received Sum RPC: %v", in)
	first := in.FirstNumber
	second := in.SecondNumber
	res := calculatorpb.SumResponse{
		Result: first + second,
	}

	return &res, nil
}

func (se *server) Multiple(ctx context.Context, in *calculatorpb.MultipleRequest) (*calculatorpb.MultipleResponse, error) {
	first := in.FirstNumber
	second := in.SecondNumber
	res := calculatorpb.MultipleResponse{
		MultipleResult: first * second,
	}
	return &res, nil
}

func (*server) ComputedAvrage(stream calculatorpb.CalculatorService_ComputedAvrageServer) error {
	result := float64(1)
	counter := 0
	for {
		counter++
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.ComputedAvrageResponse{Result: result / float64(counter)})
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		result += float64(req.Number)
	}
}

func (*server) ShapeManyTimes(req *calculatorpb.ShapeManyTimesRequest, stream calculatorpb.CalculatorService_ShapeManyTimesServer) error {
	log.Println("Hi ShapeManyTimes from server!")
	radius := req.GetRadius()

	for i := 1; i <= 10; i++ {
		res := calculatorpb.ShapeManyTimesResponse{
			Result: radius * int32(i),
		}
		stream.Send(&res)
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

func (*server) FindMaxNumEveryone(stream calculatorpb.CalculatorService_FindMaxNumEveryoneServer) error {
	log.Println("Hi FindMaxNumEveryone from server!")
	for {
		req, err := stream.Recv()
		max := int32(0)

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalln(err.Error())
			return nil
		}
		for _, num := range req.Number {
			if num > max {
				max = num
			}
		}
		stream.Send(&calculatorpb.FindMaxNumEveryoneResponse{
			Result: float64(max),
		})
	}
}

func main() {
	fmt.Println("HI im server.")
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln(err.Error())
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err = s.Serve(lis); err != nil {
		log.Fatalln(err.Error())
	}

}
