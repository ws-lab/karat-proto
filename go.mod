module github.com/ws-lab/karat-proto

require (
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/jteeuwen/go-bindata v3.0.7+incompatible // indirect
	github.com/rakyll/statik v0.1.7
	github.com/ws-lab/karat-proto/pkg/karatproto v0.0.0
	github.com/ws-lab/karat-proto/pkg/pb v0.0.0
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9 // indirect
	golang.org/x/sys v0.0.0-20200620081246-981b61492c35 // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/appengine v1.6.6 // indirect

	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.29.1
)

replace github.com/golang-migrate/migrate/v4 => github.com/golang-migrate/migrate/v4 v4.2.5

replace github.com/ws-lab/karat-proto/pkg/pb => ./pkg/pb

replace github.com/ws-lab/karat-proto/pkg/karatproto => ./pkg/karatproto

//replace google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20200619004808-3e7fca5c55db
