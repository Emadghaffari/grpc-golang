package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/dev/greet/greetpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Println("Greet function colled in server")
	firstName := req.GetGreeting().GetFirstName()
	res := &greetpb.GreetResponse{
		Result: "Hello " + firstName,
	}

	return res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	name := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		result := "Hi " + name + " invitation number: " + strconv.Itoa(i)
		res := greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(&res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("function LongGreet from server")
	result := ""
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			// return the response
			return stream.SendAndClose(
				&greetpb.LongGreetResponse{
					Result: result,
				},
			)
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		firstName := request.GetGreeting().GetFirstName()
		result += "Hi " + firstName + "!\n"
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("function GreetEveryone from server")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalln(err.Error())
			return err
		}
		firstName := req.GetGreeting().GetFirstName()
		serr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: "Hi " + firstName + "!\n",
		})
		if serr != nil {
			log.Fatalln(serr.Error())
			return serr
		}
	}
}

func (*server) SqrtRoot(ctx context.Context, req *greetpb.SqrtRootRequest) (*greetpb.SqrtRootResponse, error) {
	fmt.Println("Greet function colled in server")
	num := req.GetNumber()
	if num <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("The number is less than 0 num: %v", num))
	}
	res := &greetpb.SqrtRootResponse{
		Result: math.Sqrt(float64(num)),
	}

	return res, nil
}

func (*server) GreetWithDeadLine(ctx context.Context, req *greetpb.GreetWithDeadLineRequest) (*greetpb.GreetWithDeadLineResponse, error) {
	fmt.Println("GreetWithDeadLine function colled in server")
	fmt.Println("sleep...")

	for i := 0; i < 3; i++ {
		fmt.Println("sleep...")
		if ctx.Err() == context.Canceled {
			fmt.Println("canceld from client")
			return nil, status.Errorf(codes.Canceled, "canceld from client")
		}
		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()
	res := &greetpb.GreetWithDeadLineResponse{
		Result: "Hello " + firstName,
	}

	return res, nil
}

func main() {
	fmt.Println("Hi im server.")
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln(err.Error())
	}
	certFile := "/go/src/app/ssl/server.crt"
	keyFile := "/go/src/app/ssl/server.pem"
	creds, serverError := credentials.NewServerTLSFromFile(certFile, keyFile)
	if serverError != nil {
		log.Fatalln(err.Error())
	}

	s := grpc.NewServer(grpc.Creds(creds))

	greetpb.RegisterGreetServiceServer(s, &server{})

	if err = s.Serve(lis); err != nil {
		log.Fatalln(err.Error())
	}

}
