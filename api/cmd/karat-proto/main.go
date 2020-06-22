package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	s "strings"
	"sync"

	"github.com/ws-lab/karat-proto/pkg/karatproto"
	"github.com/ws-lab/karat-proto/pkg/pb"
	_ "github.com/ws-lab/karat-proto/statik"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:8081", "gRPC server endpoint")
	system_path        string
	database_path      string
)

func main() {
	var err error
	system_path, err = getPath()
	if err != nil {
		log.Fatal(err)
	}
	karatproto.SYSTEM_PATH = system_path
	karatproto.BASE_PATH = system_path + "/database/base.db"
	karatproto.InitDB()
	//go startGRPC()
	go startGRPC()
	//go startHTTP()
	go startHTTP()
	// Block forever
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
func getPath() (dir string, err error) {
	dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
		return
	}
	dir = s.Replace(dir, "/api/cmd/karat-proto", "", -1)
	dir = s.Replace(dir, "/build", "", -1)
	return
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, system_path+"/www/swagger.json")
}

func startGRPC() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s := grpc.NewServer()
	srv := &karatproto.GRPCServer{}
	pb.RegisterKaratProtoServer(s, srv)
	l, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("GRPC server ready...")
	log.Println("Serving GRPC port: 8081")
	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}
}

func startHTTP() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Connect to the GRPC server
	conn, err := grpc.Dial(":8081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// Register grpc-gateway
	rmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = pb.RegisterKaratProtoHandlerFromEndpoint(ctx, rmux, *grpcServerEndpoint, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Serve the swagger,
	mux := http.NewServeMux()
	mux.Handle("/", rmux)

	mux.HandleFunc("/swagger.json", serveSwagger)
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}
	staticServer := http.FileServer(statikFS)
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui", staticServer))

	log.Println("REST server ready...")
	log.Println("Serving Swagger at: http://localhost:8085/swagger-ui/")
	err = http.ListenAndServe(":8085", mux)
	if err != nil {
		log.Fatal(err)
	}
}
