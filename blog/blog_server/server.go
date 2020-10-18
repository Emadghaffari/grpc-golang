package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/dev/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var blogCollection *mongo.Collection
var notifCollection *mongo.Collection

type server struct{}
type blogItem struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID int32              `bson:"user_id"`
	Title  string             `bson:"title"`
	Tags   []string           `bson:"tags"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Println("submit a new request....")

	blog := req.GetBlog()
	data := blogItem{
		UserID: blog.GetUserId(),
		Title:  blog.GetTitle(),
		Tags:   blog.GetTags(),
	}
	res, err := blogCollection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in insert data to mongodb: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:     oid.Hex(),
			UserId: blog.GetUserId(),
			Title:  blog.GetTitle(),
			Tags:   blog.GetTags(),
		},
	}, nil
}

func (*server) CreateNotifCode(ctx context.Context, req *blogpb.CreateNotifCodeRequest) (*blogpb.CreateNotifCodeResponse, error) {
	log.Println("create a notif code....")

	rand.Seed(time.Now().UnixNano())
	phone := req.GetBlog().GetPhone()
	code := rand.Intn(999999-100000+1) + 100000
	res, err := notifCollection.InsertOne(context.Background(), bson.D{{
		"code",
		code,
	}, {
		"phone",
		phone,
	}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in insert data to mongodb: %v", err))
	}
	_, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}

	return &blogpb.CreateNotifCodeResponse{
		Result: "generated code for " + phone + "is: " + strconv.Itoa(code),
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Println("search for a blog....")
	data := &blogItem{}
	blogID := req.BlogId

	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	filter := bson.D{{"_id", oid}}
	res := blogCollection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}

	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:     data.ID.Hex(),
			UserId: data.UserID,
			Title:  data.Title,
			Tags:   data.Tags,
		},
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Println("update a blog....")
	data := &blogItem{}
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	filter := bson.D{{"_id", oid}}
	res := blogCollection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	data.Tags = blog.GetTags()
	data.Title = blog.GetTitle()
	data.UserID = blog.GetUserId()

	_, err = blogCollection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in update a value"))
	}

	return &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:     data.ID.Hex(),
			UserId: data.UserID,
			Title:  data.Title,
			Tags:   data.Tags,
		},
	}, nil

}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	log.Println("list a blog....")

	cur, err := blogCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("internal err for cur: %v", err))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("internal err: %v", err))
		}
		stream.Send(&blogpb.ListBlogResponse{Blog: toblogpb(data)})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("internal err: %v", err))
	}
	return nil
}

func main() {
	// if go code crashed we get error and line
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// start mongoDB
	client := mongodb()
	defer client.Disconnect(context.TODO())

	// start Blog server
	fmt.Println("Blog Service")
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer lis.Close()

	// certFile := "/go/src/app/ssl/server.crt"
	// keyFile := "/go/src/app/ssl/server.pem"
	// creds, serverError := credentials.NewServerTLSFromFile(certFile, keyFile)
	// if serverError != nil {
	// 	log.Fatalln(err.Error())
	// }

	// s := grpc.NewServer(grpc.Creds(creds))
	s := grpc.NewServer()
	defer s.Stop()
	blogpb.RegisterBlogServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalln(err.Error())
		}
	}()

	//wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until Control C
	<-ch
	fmt.Println("server stoppped.")
	fmt.Println("closed listen to server.")
	fmt.Println("mongodb disconnected")
	fmt.Println("Exit.")

}

func mongodb() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@mongo:27017"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalln(err.Error())
	}

	blogCollection = client.Database("blog").Collection("users")
	notifCollection = client.Database("blog").Collection("notifications")

	fmt.Println("mongoDB connected")

	return client
}

func toblogpb(item *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:     item.ID.Hex(),
		UserId: item.UserID,
		Title:  item.Title,
		Tags:   item.Tags,
	}
}
