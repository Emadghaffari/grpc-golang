package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dev/blog/blogpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hi im Client!")
	// certFile := "/go/src/app/ssl/ca.crt"
	// creds, _ := credentials.NewClientTLSFromFile(certFile, "")
	// conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("Error in grpc Dial: %v", err)
	}
	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)
	// doUnary(c)
	// readBlog(c)
	// updateBlog(c)
	// generateNotif(c)
	listofBlogs(c)

}

func doUnary(c blogpb.BlogServiceClient) {
	fmt.Println("Starting doUnary Call")
	blog := blogpb.Blog{
		UserId: 5,
		Title:  "Hi from me to all",
		Tags:   []string{"Best", "Star"},
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: &blog})
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println("Result: ", res.Blog)

}

func readBlog(c blogpb.BlogServiceClient) {
	fmt.Println("********Starting ReadBlog Call*********")
	blog := blogpb.Blog{
		UserId: 5,
		Title:  "Hi from me to all",
		Tags:   []string{"Best", "Star"},
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: &blog})
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "asd4as984da9s8d"})
	if err != nil {
		log.Printf("Blog not found: %v", err)
	}

	BlogID := res.GetBlog().GetId()
	result, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: BlogID})
	if err != nil {
		log.Printf("Blog not found: %v", err)
	}
	fmt.Println("Result: ", result)
	fmt.Println("********Done ReadBlog Call*********")

}
func updateBlog(c blogpb.BlogServiceClient) {
	blog := blogpb.Blog{
		UserId: 5,
		Title:  "Hi from me to all",
		Tags:   []string{"Best", "Star"},
	}

	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: &blog})
	if err != nil {
		log.Fatalln(err.Error())
	}

	BlogID := res.GetBlog().GetId()
	blogUP := blogpb.Blog{
		Id:     BlogID,
		UserId: 5,
		Title:  "Hi from me to all UPDATED",
		Tags:   []string{"ok", "updated"},
	}
	ures, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: &blogUP})
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println("Result: ", ures.Blog)

}

func generateNotif(c blogpb.BlogServiceClient) {
	fmt.Println("Starting generateNotif Call")
	blog := blogpb.Blog{
		UserId: 5,
		Title:  "Hi from me to all",
		Tags:   []string{"Best", "Star"},
		Phone:  "09355960597",
	}

	res, err := c.CreateNotifCode(context.Background(), &blogpb.CreateNotifCodeRequest{Blog: &blog})
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Println("Result: ", res.Result)
}

func listofBlogs(c blogpb.BlogServiceClient) {
	resStream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
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
		time.Sleep(time.Millisecond * 1000)
		log.Printf("Result for list Blog: %v", res.GetBlog())
	}
}
