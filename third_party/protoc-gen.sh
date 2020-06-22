#!/bin/bash
export PATH=$PATH:$GOPATH/bin
mkdir -p build api/swagger
 protoc -I api/proto  --proto_path=third_party --go_out=plugins=grpc:pkg/pb api/proto/grpcadder.proto
 protoc -I api/proto  --proto_path=third_party --grpc-gateway_out=logtostderr=true:pkg/pb api/proto/grpcadder.proto
 protoc -I api/proto  --proto_path=third_party --swagger_out=logtostderr=true:api/swagger api/proto/grpcadder.proto
cp api/swagger/grpcadder.swagger.json www/swagger.json
statik -src=www/
