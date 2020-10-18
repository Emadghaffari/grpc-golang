package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dev/webspice/webspicepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	layoutISO = "2006-01-02"
	layoutUS  = "January 2, 2006"
)

func main() {
	fmt.Println("HI Client.")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	c := webspicepb.NewWebspiceClient(conn)
	// unary(c)
	// clientStreaming(c)
	// serversiceStreaming(c)
	// bidi(c)
	// _, err = createUser(c)
	// if err != nil {
	// 	log.Fatalln("Error in Create new User: ", err)
	// }
	// searched, err := searchUser(c, created)
	// if err != nil {
	// 	log.Fatalln("Error in Create new User: ", err)
	// }
	// updated, err := updateUser(c, searched)
	// if err != nil {
	// 	log.Fatalln("Error in Create new User: ", err)
	// }
	// deleteUser(c, updated)

	// usersMany(c)
	searchEvery(c)
}

func unary(c webspicepb.WebspiceClient) {
	fmt.Println("Starting unary Call")
	req := webspicepb.WebspiceRequest{
		FirstName: "Emad",
		Age:       25,
	}

	res, err := c.Webspice(context.Background(), &req)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Printf("The response from server: %v \n", res.Result)
}

func clientStreaming(c webspicepb.WebspiceClient) {
	fmt.Println("Starting clientStreaming Call")
	req := webspicepb.WebspiceManyRequest{
		FirstName: "Emad",
		Age:       25,
	}

	resq, err := c.WebspiceMany(context.Background(), &req)
	if err != nil {
		log.Fatalln(err.Error())
	}
	for {
		res, err := resq.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		fmt.Printf("The response from server: %v \n", res.Result)
	}
	fmt.Printf("DONE.")
}

func serversiceStreaming(c webspicepb.WebspiceClient) {
	fmt.Println("Starting serversiceStreaming Call")
	reqs := []*webspicepb.WebspiceLongRequest{
		&webspicepb.WebspiceLongRequest{
			FirstName: "Emad",
			Age:       25,
		},
		&webspicepb.WebspiceLongRequest{
			FirstName: "atena",
			Age:       18,
		},
	}
	stream, err := c.WebspiceLong(context.Background())
	if err != nil {
		log.Fatalln(err.Error())
	}
	for _, req := range reqs {
		fmt.Printf("sending request ... %v \n", req)
		stream.Send(req)
		time.Sleep(time.Millisecond * 1000)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Printf("Response for ClientSideStreaming:\n%v \n", res.Result)

	fmt.Printf("DONE.")
}

func bidi(c webspicepb.WebspiceClient) {
	fmt.Println("Starting bidi Call")
	waitc := make(chan struct{})
	reqs := []*webspicepb.WebspiceEveryDomainRequest{
		&webspicepb.WebspiceEveryDomainRequest{
			Domain: ".ir",
		},
		&webspicepb.WebspiceEveryDomainRequest{
			Domain: ".com",
		},
		&webspicepb.WebspiceEveryDomainRequest{
			Domain: ".org",
		},
		&webspicepb.WebspiceEveryDomainRequest{
			Domain: ".me",
		},
		&webspicepb.WebspiceEveryDomainRequest{
			Domain: ".info",
		},
	}
	stream, err := c.WebspiceEveryDomain(context.Background())
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
	}

	go func() {
		for _, req := range reqs {
			fmt.Printf("send a new request:%v\n", req)
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
			}
			log.Printf("response from server: %v", res.Result)
		}
	}()

	<-waitc
	fmt.Printf("DONE.")
}

func searchEvery(c webspicepb.WebspiceClient) {
	fmt.Println("start searchEvery")
	waitc := make(chan struct{})
	reqs := []*webspicepb.SearchEveryUserRequest{
		&webspicepb.SearchEveryUserRequest{
			User: &webspicepb.User{
				Id: "5f795f771f53f2206acd6fcf",
			},
		},
		&webspicepb.SearchEveryUserRequest{
			User: &webspicepb.User{
				Id: "5f795f821f53f2206acd6fd0",
			},
		},
		&webspicepb.SearchEveryUserRequest{
			User: &webspicepb.User{
				Id: "5f798eaa60325437128771f6",
			},
		},
		&webspicepb.SearchEveryUserRequest{
			User: &webspicepb.User{
				Id: "5f798ef560325437128771f8",
			},
		},
		&webspicepb.SearchEveryUserRequest{
			User: &webspicepb.User{
				Id: "5f798f48b44b70cd73ebde24",
			},
		},
	}

	stream, err := c.SearchEveryUser(context.Background())
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
	}

	go func() {
		for _, req := range reqs {
			fmt.Println("Send new userID")
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
			}
			log.Printf("response from server: %v", res.User)
		}
	}()
	<-waitc
	log.Printf("DONE")

}

func createUser(c webspicepb.WebspiceClient) (*webspicepb.CreateUserResponse, error) {
	fmt.Println("--------------create a new user--------------")
	t := time.Now()
	req := webspicepb.CreateUserRequest{
		User: &webspicepb.User{
			FirstName:   "Emad",
			LastName:    "Ghaffari",
			Phone:       "+9355960597",
			Permissions: []string{"superadmin"},
			Status:      webspicepb.User_DEACTIVE,
			CreatedAt:   t.Format(layoutISO),
		},
	}

	res, err := c.CreateUser(context.Background(), &req)
	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			log.Printf("Error code: %v\n", resErr.Code())
			log.Printf("Error Message: %v\n", resErr.Message())
		} else {
			log.Fatalln("The Big Error framework, ", err)
		}
		return nil, err
	}
	fmt.Printf("create- The response from server:\n %v \n", res.User)
	return res, nil
}

func searchUser(c webspicepb.WebspiceClient, req *webspicepb.CreateUserResponse) (*webspicepb.SearchUserResponse, error) {
	fmt.Println("--------------search a user--------------")
	user := webspicepb.SearchUserRequest{
		User: &webspicepb.User{
			Id: req.GetUser().GetId(),
		},
	}
	res, err := c.SearchUser(context.Background(), &user)
	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			log.Printf("Error code: %v\n", resErr.Code())
			log.Printf("Error Message: %v\n", resErr.Message())
		} else {
			log.Fatalln("The Big Error framework, ", err)
		}
		return nil, err
	}
	fmt.Printf("search- The response from server:\n %v \n", res.User)
	return res, nil
}

func updateUser(c webspicepb.WebspiceClient, req *webspicepb.SearchUserResponse) (*webspicepb.UpdateUserResponse, error) {
	fmt.Println("--------------update a old user--------------")
	user := webspicepb.UpdateUserRequest{
		User: &webspicepb.User{
			Id:          req.GetUser().GetId(),
			FirstName:   "Emad.",
			LastName:    "Ghaffari",
			Phone:       "09355960597",
			Permissions: req.GetUser().GetPermissions(),
		},
	}
	res, err := c.UpdateUser(context.Background(), &user)
	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			log.Printf("Error code: %v\n", resErr.Code())
			log.Printf("Error Message: %v\n", resErr.Message())
		} else {
			log.Fatalln("The Big Error framework, ", err)
		}
		return nil, err
	}
	fmt.Printf("update- The response from server:(old-user)\n %v \n", req.User)
	fmt.Printf("update- The response from server:(new-user)\n %v \n", res.User)
	return res, nil
}

func deleteUser(c webspicepb.WebspiceClient, req *webspicepb.UpdateUserResponse) {
	fmt.Println("--------------update a old user--------------")
	user := webspicepb.DeleteUserRequest{
		User: &webspicepb.User{
			Id: req.GetUser().GetId(),
		},
	}
	res, err := c.DeleteUser(context.Background(), &user)
	if err != nil {
		resErr, ok := status.FromError(err)
		if ok {
			log.Printf("Error code: %v\n", resErr.Code())
			log.Printf("Error Message: %v\n", resErr.Message())
		} else {
			log.Fatalln("The Big Error framework, ", err)
		}
		return
	}
	fmt.Printf("delete The response from server:\n %v \n", res.User)
}

func usersMany(c webspicepb.WebspiceClient) {
	fmt.Println("Start usersMany")
	req := &webspicepb.SearchUserManyRequest{
		User: &webspicepb.User{},
	}

	ctx, err := c.SearchUserMany(context.Background(), req)
	if err != nil {
		serr, ok := status.FromError(err)
		if ok {
			fmt.Println(serr.Code())
			fmt.Println(serr.Message())
		} else {
			log.Println("BIG Error")
		}
	}
	for {
		res, err := ctx.Recv()
		time.Sleep(1000 * time.Millisecond)

		if err == io.EOF {
			return
		}
		if err != nil {
			serr, ok := status.FromError(err)
			if ok {
				fmt.Println(serr.Code())
				fmt.Println(serr.Message())
			} else {
				log.Println("BIG Error")
			}
		}
		fmt.Println("Result: ", res.User)
	}
}
