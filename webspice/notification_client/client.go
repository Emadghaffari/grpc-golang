package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dev/webspice/notificationpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	layoutISO = "2006-01-02 15:04 MST"
	layoutUS  = "January 2, 2006"
)

func main() {
	fmt.Println("HI Client.")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer conn.Close()

	c := notificationpb.NewWebspiceClient(conn)

	_, err = createnotif(c)
	if err != nil {
		log.Fatalln("Error in Create new User: ", err)
	}
	// updated, err := updateUser(c, searched)
	// if err != nil {
	// 	log.Fatalln("Error in Create new User: ", err)
	// }
}

func createnotif(c notificationpb.WebspiceClient) (*notificationpb.CreateNotificationResponse, error) {
	fmt.Println("--------------create a new user--------------")
	req := notificationpb.CreateNotificationRequest{
		Notification: &notificationpb.Notification{
			For:   "Register Account",
			Phone: "+9355960597",
		},
	}

	res, err := c.CreateNotification(context.Background(), &req)
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
	fmt.Printf("create- The response from server:\n %v \n", res.Notification)
	return res, nil
}

// func updateUser(c notificationpb.WebspiceClient, req *notificationpb.SearchUserResponse) (*notificationpb.UpdateUserResponse, error) {
// 	fmt.Println("--------------update a old user--------------")
// 	user := notificationpb.UpdateUserRequest{
// 		User: &notificationpb.User{
// 			Id:          req.GetUser().GetId(),
// 			FirstName:   "Emad.",
// 			LastName:    "Ghaffari",
// 			Phone:       "09355960597",
// 			Permissions: req.GetUser().GetPermissions(),
// 		},
// 	}
// 	res, err := c.UpdateUser(context.Background(), &user)
// 	if err != nil {
// 		resErr, ok := status.FromError(err)
// 		if ok {
// 			log.Printf("Error code: %v\n", resErr.Code())
// 			log.Printf("Error Message: %v\n", resErr.Message())
// 		} else {
// 			log.Fatalln("The Big Error framework, ", err)
// 		}
// 		return nil, err
// 	}
// 	fmt.Printf("update- The response from server:(old-user)\n %v \n", req.User)
// 	fmt.Printf("update- The response from server:(new-user)\n %v \n", res.User)
// 	return res, nil
// }
