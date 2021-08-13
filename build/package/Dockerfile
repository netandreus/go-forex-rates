# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:1.16.6

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
# Install system packages
RUN apt-get update \
    &&  apt-get -y install \
        mc

# Add Maintainer Info
LABEL maintainer="Andrey <netandreus@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build swagger scecifications
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN swag init --output=./api

# Build the Go app
RUN go build -o main .

# Expose port 9090 to the outside world
EXPOSE 9090

# Run the executable
CMD ["./main"]