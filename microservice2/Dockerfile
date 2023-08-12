FROM golang:1.21.0

# Setting the working directory in the container
WORKDIR /app

# Copying the Go modules list and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copying application files
COPY . .

# Building executable
RUN go build -o main .

# Exposing port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
