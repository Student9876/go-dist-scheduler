# 1. Builder 
FROM golang:1.24-alpine AS builder

# set the folder inside the container where we'll work 
WORKDIR /app

# copy dependency files first as this helps docker to cache dependencies 
COPY go.mod go.sum ./
RUN go mod download

# copy rest of the source code
COPY . .

# build the application
# we name the output binary main
RUN go build -o main ./cmd/api/main.go

# 2. The runner (runs the code)
FROM alpine:latest

# set working directory of the final image
WORKDIR /root/

# copy the binary from the builder stage
COPY --from=builder /app/main .

# The command to run when the container starts
CMD ["./main"]