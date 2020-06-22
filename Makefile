.PHONY: build clean serve pkg/migrations generate_static_swagger dev-requirements
generate_static_swagger:
		sh ./third_party/protoc-gen.sh

build_full: pkg/migrations generate_static_swagger
	go build -o build/karat-proto api/cmd/karat-proto/main.go

build: dev-requirements pkg/migrations
	go build -o build/karat-proto api/cmd/karat-proto/main.go
	
clean:
	@echo "Cleaning up workspace"
	@rm -rf build statik database pkg/migrations/migrations_gen.go
	@rm -f  api/cmd/karat-proto/karat-proto
    		
pkg/migrations:
	@echo "Generating static files"
	@go generate pkg/migrations/migrations.go

dev-requirements: 
	go mod download
	go get -u github.com/jteeuwen/go-bindata
	go get -u github.com/rakyll/statik
	go get -u github.com/golang/protobuf
	
	go install github.com/jteeuwen/go-bindata/go-bindata
	go install github.com/rakyll/statik
	go install google.golang.org/protobuf/cmd/protoc-gen-go
		
	statik -src=www/
	
serve: build
	@echo "Starting Karat Proto Server"
	./build/karat-proto