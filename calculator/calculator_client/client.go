package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dev/calculator/calculatorpb"
	"google.golang.org/grpc"
)

var wg sync.WaitGroup

func main() {
	fmt.Println("Hi im Client!")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	c := calculatorpb.NewCalculatorServiceClient(conn)
	// wg.Add(3)
	// go doUnary(c)
	// go doServerStream(c)
	// go doClientStream(c)
	// wg.Wait()
	doBIDI(c)

}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	defer wg.Done()
	fmt.Println("Starting calculatorpb doUnary Call")
	sumreq := calculatorpb.SumRequest{
		FirstNumber:  10,
		SecondNumber: 25,
	}
	mulreq := calculatorpb.MultipleRequest{
		FirstNumber:  10,
		SecondNumber: 25,
	}

	res, err := c.Sum(context.Background(), &sumreq)
	if err != nil {
		log.Fatalln(err.Error())
	}

	res2, err := c.Multiple(context.Background(), &mulreq)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println("Sum Result: ", res.Result)
	fmt.Println("Mul Result: ", res2.MultipleResult)
}
func doServerStream(c calculatorpb.CalculatorServiceClient) {
	defer wg.Done()
	req := calculatorpb.ShapeManyTimesRequest{
		Radius: 3,
	}
	resStream, err := c.ShapeManyTimes(context.Background(), &req)
	if err != nil {
		log.Fatalln(err.Error())
	}
	for {
		res, err := resStream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Printf("Result for Circle Radius: %v", res.Result)
	}
}

func doClientStream(c calculatorpb.CalculatorServiceClient) {
	defer wg.Done()
	fmt.Println("start ClientStream!")
	reqs := []*calculatorpb.ComputedAvrageRequest{
		&calculatorpb.ComputedAvrageRequest{
			Number: 10,
		},
		&calculatorpb.ComputedAvrageRequest{
			Number: 7,
		},
		&calculatorpb.ComputedAvrageRequest{
			Number: 3,
		},
		&calculatorpb.ComputedAvrageRequest{
			Number: 15,
		},
		&calculatorpb.ComputedAvrageRequest{
			Number: 4,
		},
		&calculatorpb.ComputedAvrageRequest{
			Number: 1,
		},
	}
	context, err := c.ComputedAvrage(context.Background())
	if err != nil {
		log.Fatalln(err.Error())
	}
	for _, req := range reqs {
		err := context.Send(req)
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Printf("Push a number for ComputedAvrage: %v", req.Number)
		time.Sleep(time.Millisecond * 1000)
	}
	resp, err := context.CloseAndRecv()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("Result for ComputedAvrage: %v", resp.Result)
}

func doBIDI(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("start doBIDI!")
	waitc := make(chan struct{})
	requests := []*calculatorpb.FindMaxNumEveryoneRequest{
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{10, 7, 2, 9, 1, -20, 12},
		},
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{7, 7, 2, 9, 1, -20, 4},
		},
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{3, 7, 2, 9, 1, -20, 50},
		},
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{15, 7, 2, 9, 1, -20, 14},
		},
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{4, 7, 2, 9, 1, -20, 99},
		},
		&calculatorpb.FindMaxNumEveryoneRequest{
			Number: []int32{1, 7, 2, 9, 1, -20, 0},
		},
	}
	stream, err := c.FindMaxNumEveryone(context.Background())
	if err != nil {
		log.Fatalln(err.Error())
	}

	go func() {
		for _, req := range requests {
			log.Printf("send request: %v", req.Number)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				close(waitc)
				log.Fatalln(err.Error())
				break
			}
			log.Printf("Result for Circle Radius: %v", res.Result)
		}
	}()

	<-waitc
}
