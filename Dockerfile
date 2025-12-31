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
RUN go build -o api_bin ./cmd/api/main.go
# build scheduler
RUN go build -o scheduler_bin ./cmd/scheduler/main.go
# build worker 
RUN go build -o worker_bin ./cmd/worker/main.go

# 2. The runner (runs the code)
FROM alpine:latest

# set working directory of the final image
WORKDIR /root/

# copy the binary from the builder stage
COPY --from=builder /app/api_bin .
COPY --from=builder /app/scheduler_bin .
COPY --from=builder /app/worker_bin .

# The command to run when the container starts
