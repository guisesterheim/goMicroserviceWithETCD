FROM golang:latest

RUN mkdir /app
ADD . /app/

# Move to working directory
WORKDIR /app

# Get dependencies
RUN go get "github.com/gorilla/mux"
RUN go get "go.etcd.io/etcd/clientv3"

# Build the application
RUN go build -o go-homework.go .

# Expose necessary port
EXPOSE 8080

# Command to run when starting the container
CMD ["/app/go-homework.go"]