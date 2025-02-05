FROM docker.io/golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o aws-key-rotation .

#-----

FROM docker.io/alpine:latest

WORKDIR /root/

COPY --from=builder /app/aws-key-rotation .

CMD ["./aws-key-rotation"]
