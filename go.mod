module github.com/ws-lab/karat-proto

require (
	cloud.google.com/go v0.72.0 // indirect
	cloud.google.com/go/storage v1.12.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403 // indirect
	github.com/creack/pty v1.1.11 // indirect
	github.com/envoyproxy/go-control-plane v0.9.7 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/pprof v0.0.0-20201117184057-ae444373da19 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/iancoleman/strcase v0.1.2 // indirect
	github.com/jteeuwen/go-bindata v3.0.7+incompatible // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lyft/protoc-gen-star v0.5.2 // indirect
	github.com/pkg/sftp v1.12.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/spf13/afero v1.4.1 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/ws-lab/karat-proto/pkg/karatproto v0.0.0
	github.com/ws-lab/karat-proto/pkg/pb v0.0.0
	golang.org/x/crypto v0.0.0-20201124201722-c8d3bf9c5392 // indirect
	golang.org/x/mod v0.3.1-0.20200828183125-ce943fd02449 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58 // indirect
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20201201161351-ac6f37ff4c2a // indirect
	google.golang.org/appengine v1.6.7 // indirect

	google.golang.org/genproto v0.0.0-20201201144952-b05cb90ed32e
	google.golang.org/grpc v1.33.2
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/golang-migrate/migrate/v4 => github.com/golang-migrate/migrate/v4 v4.2.5

replace github.com/ws-lab/karat-proto/pkg/pb => ./pkg/pb

replace github.com/ws-lab/karat-proto/pkg/karatproto => ./pkg/karatproto

//replace google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20200619004808-3e7fca5c55db
