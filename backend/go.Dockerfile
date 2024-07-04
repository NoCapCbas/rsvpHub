# Use the specific Go version Alpine image
FROM golang:1.22.1-alpine3.19

# Set the working directory inside the container
WORKDIR /backend

# Copy the entire current directory to the container's working directory
COPY . .

# Download and install dependencies using go mod tidy
RUN go get -d -v ./...

# Build the Go app
RUN go build -o main

# Expose port 8000 to the outside world
EXPOSE 8001

# Command to run the executable
CMD ["./main"]
