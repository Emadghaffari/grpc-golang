#!/bin/sh

# greet
protoc greet/greetpb/greet.proto  --go_out=plugins=grpc:.

# calculator
protoc calculator/calculatorpb/calculator.proto  --go_out=plugins=grpc:.

# webspice
protoc webspice/webspicepb/webspice.proto  --go_out=plugins=grpc:.
protoc webspice/notificationpb/notification.proto  --go_out=plugins=grpc:.

# blog
protoc blog/blogpb/blog.proto  --go_out=plugins=grpc:.