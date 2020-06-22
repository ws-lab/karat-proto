FROM golang:1.11-alpine AS development

ENV PROJECT_PATH=/karat-proto
ENV PATH=$PATH:$PROJECT_PATH/build
ENV PATH=$PATH:$GOPATH/bin
ENV GO111MODULE=on

RUN apk add --update --no-cache make ca-certificates git bash gcc libc-dev
RUN mkdir -p $PROJECT_PATH
WORKDIR $PROJECT_PATH

COPY go.mod .
COPY go.sum .
COPY pkg/karatproto/go.sum .
COPY pkg/karatproto/go.mod .
COPY pkg/pb/go.sum .
COPY pkg/pb/go.mod .

#RUN go mod download
COPY . .
RUN make dev-requirements
RUN make pkg/migrations
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/karat-proto $PROJECT_PATH/api/cmd/karat-proto/main.go

## A better scratch
FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates bash
COPY --from=development /go/bin/karat-proto /go/bin/karat-proto

EXPOSE 8085
EXPOSE 8081

CMD ["/go/bin/karat-proto"]
