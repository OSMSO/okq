# golang image where workspace (GOPATH) configured at /go.
FROM golang:1.10.3

# Install dependencies
ADD ./vendor/github.com/HuKeping /go/src/github.com/HuKeping
ADD ./vendor/github.com/alex023 /go/src/github.com/alex023
ADD ./vendor/github.com/gorilla /go/src/github.com/gorilla
ADD ./vendor/github.com/go-sql-driver /go/src/github.com/go-sql-driver
ADD ./vendor/github.com/jinzhu /go/src/github.com/jinzhu

# copy the local package files to the container workspace
ADD . /go/src/github.com/osmso/clock

# Setting up working directory
WORKDIR /go/src/github.com/osmso/clock

# Build the clock command inside the container.
RUN go install github.com/osmso/clock

# Run the clock microservice when the container starts.
ENTRYPOINT /go/bin/clock

# Service listens on port 3000.
EXPOSE 3000