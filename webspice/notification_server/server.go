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

	"github.com/dev/webspice/notificationpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

var database *mongo.Database

type notification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Phone        string             `bson:"phone"`
	For          string             `bson:"for"`
	Code         string             `bson:"code"`
	CreatedAt    string             `bson:"created_at"`
	DeprecatedAt string             `bson:"deprecated_at"`
	Status       string             `bson:"status"`
}

const (
	layoutISO = "2006-01-02 15:04"
	layoutUS  = "January 2, 2006"
)

type server struct{}

func (*server) CreateNotification(ctx context.Context, req *notificationpb.CreateNotificationRequest) (*notificationpb.CreateNotificationResponse, error) {
	fmt.Println("CreateNotification.")
	rand.Seed(time.Now().UnixNano())
	loc, _ := time.LoadLocation("UTC")
	t := time.Now().In(loc)
	code := rand.Intn(999999-100000+1) + 100000

	data := &notificationpb.Notification{
		Phone:        req.GetNotification().GetPhone(),
		For:          req.GetNotification().GetFor(),
		Code:         strconv.Itoa(code),
		CreatedAt:    t.Format(layoutISO),
		DeprecatedAt: t.Add(time.Minute * 5).Format(layoutISO),
		Status:       notificationpb.Notification_DEACTIVE,
	}

	cres, err := database.Collection("notifs").InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in insert data to mongodb: %v", err))
	}
	oid, ok := cres.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get OID from InsertedID: %v", oid))
	}
	data.Id = oid.Hex()
	res := &notificationpb.CreateNotificationResponse{
		Notification: data,
	}
	return res, nil
}

func (*server) UpdateNotification(ctx context.Context, req *notificationpb.UpdateNotificationRequest) (*notificationpb.UpdateNotificationResponse, error) {
	fmt.Println("UpdateNotification.")

	data := &notificationpb.Notification{}
	filter := bson.D{{"phone", req.GetNotification().GetPhone()}}
	res := database.Collection("notifs").FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in search data from mongodb: %v", err))
	}
	data.Status = notificationpb.Notification_ACTIVE
	_, err := database.Collection("notifs").ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in update a value"))
	}
	response := &notificationpb.UpdateNotificationResponse{
		Notification: data,
	}
	return response, nil
}

func main() {
	// if go code crashed we get error and line
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// start mongoDB connection
	client := mongodb()
	defer client.Disconnect(context.TODO())

	// start server
	fmt.Println("Hi From server.")
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer lis.Close()

	s := grpc.NewServer()
	defer s.Stop()
	notificationpb.RegisterWebspiceServer(s, &server{})
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

	if err = client.Connect(ctx); err != nil {
		log.Fatalln(err.Error())
	}

	database = client.Database("webspice")

	fmt.Println("mongoDB connected")

	return client
}
