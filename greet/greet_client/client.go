package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dev/greet/greetpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hi im Client!")
	certFile := "/go/src/app/ssl/ca.crt"
	creds, _ := credentials.NewClientTLSFromFile(certFile, "")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)
	// doUnary(c)
	// doServerSideStreaming(c)
	// doClientSideStreaming(c)
	// bIDIstreaming(c)
	// checkError(c)
	doUnaryWithDeadline(c, 5*time.Second)
	doUnaryWithDeadline(c, 2*time.Second)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doUnary Call")
	req := greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Emad",
			LastName:  "Ghaffari",
		},
	}

	res, err := c.Greet(context.Background(), &req)
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println("Result: ", res.Result)
}

func doServerSideStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doServerSideStreaming Call")
	req := greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Emad",
			LastName:  "Ghaffari",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), &req)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Printf("Response for GreetManyTimes: %v", msg.Result)
	}

}

func doClientSideStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting doClientSideStreaming Call")
	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Emad",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Mamad",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "READ",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "logical",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "mina",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalln(err.Error())
	}
	for _, req := range requests {
		fmt.Printf("sending request ... %v \n", req)
		stream.Send(req)
		time.Sleep(time.Millisecond * 1000)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("Response for ClientSideStreaming: %v \n", res.Result)
}

func bIDIstreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting BIDIstreaming Call")
	wiatc := make(chan struct{})
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Emad",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Mamad",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "READ",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "logical",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "mina",
			},
		},
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalln(err.Error())
	}

	go func() {
		for _, req := range requests {
			fmt.Printf("sending request ... %v \n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(wiatc)
				break
			}
			if err != nil {
				log.Fatalln(err.Error())
				close(wiatc)
				break
			}
			fmt.Printf("response from server: %v\n", res.Result)
		}
	}()

	<-wiatc
}

func checkError(c greetpb.GreetServiceClient) {
	fmt.Println("Starting checkError Call")
	req := greetpb.SqrtRootRequest{
		Number: -10,
	}

	res, err := c.SqrtRoot(context.Background(), &req)
	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			log.Println(resErr.Code())
			log.Println(resErr.Message())
			if resErr.Code() == codes.InvalidArgument {
				log.Println("the InvalidArgument from client")
			}
		} else {
			log.Fatalln("The Big Error framework, ", err)
		}
	} else {
		fmt.Println("Result: ", res.Result)

	}

}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, tm time.Duration) {
	fmt.Println("Starting doUnaryWithDeadline Call")
	req := greetpb.GreetWithDeadLineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Emad",
			LastName:  "Ghaffari",
		},
	}
	ctx, close := context.WithDeadline(context.Background(), time.Now().Add(tm))
	defer close()
	res, err := c.GreetWithDeadLine(ctx, &req)
	if err != nil {
		statusError, ok := status.FromError(err)
		if ok {
			if statusError.Code() == codes.Canceled {
				fmt.Printf("TimeOut: %v", statusError)
			} else {
				fmt.Printf("unexpected error: %v ", statusError)
			}
		} else {
			fmt.Printf("bad Error RPC: %v ", statusError)
		}
		return
	}

	fmt.Println("Result: ", res.Result)
}
