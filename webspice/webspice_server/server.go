package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/dev/webspice/webspicepb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var database *mongo.Database

type user struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	FirstName   string             `bson:"first_name"`
	LastName    string             `bson:"last_name"`
	Phone       string             `bson:"phone"`
	CreatedAt   string             `bson:"created_at"`
	Permissions []string           `bson:"permissions"`
	Status      string             `bson:"status"`
}
type server struct{}

func (*server) Webspice(ctx context.Context, req *webspicepb.WebspiceRequest) (*webspicepb.WebspiceResponse, error) {
	fmt.Println("Webspice.")
	name := req.GetFirstName()
	age := req.GetAge()
	res := &webspicepb.WebspiceResponse{
		Result: "Hi " + name + " Your age is: " + strconv.Itoa(int(age)),
	}
	return res, nil
}

func (*server) WebspiceMany(req *webspicepb.WebspiceManyRequest, stream webspicepb.Webspice_WebspiceManyServer) error {
	fmt.Println("WebspiceMany.")

	name := req.GetFirstName()
	age := req.GetAge()

	for i := 0; i < 20; i++ {
		time.Sleep(250 * time.Millisecond)
		res := webspicepb.WebspiceManyResponse{
			Result: "Hi " + name + " Your age is: " + strconv.Itoa(int(age+int32(i))),
		}
		stream.Send(&res)
	}
	return nil
}

func (*server) WebspiceLong(stream webspicepb.Webspice_WebspiceLongServer) error {
	fmt.Println("WebspiceLong.")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&webspicepb.WebspiceLongResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalln(err.Error())
		}
		name := req.GetFirstName()
		age := req.GetAge()
		result += "Hi " + name + " Your age is: " + strconv.Itoa(int(age)) + "\n"
	}
}

func (*server) WebspiceEveryDomain(stream webspicepb.Webspice_WebspiceEveryDomainServer) error {
	fmt.Println("WebspiceEveryDomain.")
	result := []string{}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return status.Errorf(codes.OK, fmt.Sprintf("success"))
		}
		if err != nil {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("the error from server: %v", err.Error()))
		}
		result = append(result, req.GetDomain())
		fmt.Println("calculate and check price for this domain-> ", req.GetDomain())
		err = stream.Send(&webspicepb.WebspiceEveryDomainResponse{
			Result: result,
		})
		if err != nil {
			return status.Errorf(codes.OutOfRange, fmt.Sprintf("error in send data from server to client: %v", err.Error()))
		}
	}

}

func (*server) CreateUser(ctx context.Context, req *webspicepb.CreateUserRequest) (*webspicepb.CreateUserResponse, error) {
	log.Println("start to create a user")
	collection := database.Collection("users")
	user := user{
		FirstName:   req.GetUser().GetFirstName(),
		LastName:    req.GetUser().GetLastName(),
		Phone:       req.GetUser().GetPhone(),
		CreatedAt:   req.GetUser().GetCreatedAt(),
		Permissions: req.GetUser().GetPermissions(),
		Status:      webspicepb.User_Status_name[int32(req.GetUser().GetStatus())],
	}

	res, err := collection.InsertOne(context.Background(), &user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in Insert a new User: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get ObjectID from response: %v", err))
	}

	return &webspicepb.CreateUserResponse{
		User: &webspicepb.User{
			Id:          oid.Hex(),
			FirstName:   req.GetUser().GetFirstName(),
			LastName:    req.GetUser().GetLastName(),
			Phone:       req.GetUser().GetPhone(),
			CreatedAt:   req.GetUser().GetCreatedAt(),
			Permissions: req.GetUser().GetPermissions(),
			Status:      req.GetUser().GetStatus(),
		},
	}, nil
}

func (*server) UpdateUser(ctx context.Context, req *webspicepb.UpdateUserRequest) (*webspicepb.UpdateUserResponse, error) {
	log.Println("start to update a user")
	user := &user{}
	collection := database.Collection("users")
	userID := req.GetUser().GetId()
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	filter := bson.D{{"_id", oid}}

	search := collection.FindOne(context.Background(), filter)
	if err = search.Decode(user); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("user not found in database..."))
	}

	user.FirstName = req.GetUser().GetFirstName()
	user.LastName = req.GetUser().GetLastName()
	user.Phone = req.GetUser().GetPhone()
	user.Permissions = req.GetUser().GetPermissions()

	_, err = collection.ReplaceOne(context.Background(), filter, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("update failed: %v", err))
	}

	return &webspicepb.UpdateUserResponse{
		User: &webspicepb.User{
			Id:          user.ID.Hex(),
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Phone:       user.Phone,
			CreatedAt:   user.CreatedAt,
			Permissions: user.Permissions,
		},
	}, nil
}

func (*server) DeleteUser(ctx context.Context, req *webspicepb.DeleteUserRequest) (*webspicepb.DeleteUserResponse, error) {
	log.Println("start to search a user")
	collection := database.Collection("users")
	userID := req.GetUser().GetId()
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	filter := bson.D{{"_id", oid}}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}

	return &webspicepb.DeleteUserResponse{
		User: &webspicepb.User{
			Id:          oid.Hex(),
			FirstName:   req.GetUser().GetFirstName(),
			LastName:    req.GetUser().GetLastName(),
			Phone:       req.GetUser().GetPhone(),
			CreatedAt:   req.GetUser().GetCreatedAt(),
			Permissions: req.GetUser().GetPermissions(),
		},
	}, nil
}

func (*server) SearchUser(ctx context.Context, req *webspicepb.SearchUserRequest) (*webspicepb.SearchUserResponse, error) {
	log.Println("start to search a user")
	user := &user{}
	collection := database.Collection("users")
	userID := req.GetUser().GetId()
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error in get oid from res"))
	}
	filter := bson.D{{"_id", oid}}

	search := collection.FindOne(context.Background(), filter)
	if err = search.Decode(user); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("user not found in database..."))
	}
	return &webspicepb.SearchUserResponse{
		User: &webspicepb.User{
			Id:          user.ID.Hex(),
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Phone:       user.Phone,
			CreatedAt:   user.CreatedAt,
			Permissions: user.Permissions,
		},
	}, nil
}

func (*server) SearchUserMany(req *webspicepb.SearchUserManyRequest, stream webspicepb.Webspice_SearchUserManyServer) error {
	cur, err := database.Collection("users").Find(context.Background(), bson.D{})
	if err != nil {
		status.Errorf(codes.Internal, fmt.Sprintf("error in the Find: %v", err))
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {

		data := &user{}
		err := cur.Decode(data)
		if err != nil {
			status.Errorf(codes.Internal, fmt.Sprintf("error in the Find: %v", err))
		}
		stream.Send(&webspicepb.SearchUserManyResponse{User: &webspicepb.User{
			Id:          data.ID.Hex(),
			FirstName:   data.FirstName,
			LastName:    data.LastName,
			Phone:       data.Phone,
			CreatedAt:   data.CreatedAt,
			Permissions: data.Permissions,
		}})
	}
	return nil

}

func (*server) SearchEveryUser(stream webspicepb.Webspice_SearchEveryUserServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			return status.Errorf(codes.OK, fmt.Sprintf("success"))
		}
		if err != nil {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("the error from server: %v", err.Error()))
		}
		oid, err := primitive.ObjectIDFromHex(data.GetUser().GetId())
		filter := bson.D{{"_id", oid}}
		fmt.Println("start process for ID: ", oid)
		collection := database.Collection("users")
		res := collection.FindOne(context.Background(), filter)
		user := &user{}
		if err := res.Decode(user); err != nil {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Decode: %v", err.Error()))
		}
		stream.Send(&webspicepb.SearchEveryUserResponse{User: &webspicepb.User{
			Id:          user.ID.Hex(),
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Phone:       user.Phone,
			CreatedAt:   user.CreatedAt,
			Permissions: user.Permissions,
		}})
	}
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
	webspicepb.RegisterWebspiceServer(s, &server{})
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
