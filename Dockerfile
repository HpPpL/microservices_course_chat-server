FROM golang:1.20.3-alpine AS builder

COPY . /github.com/HpPpL/microservices_course_chat-server/grpc/source
COPY go.mod /github.com/HpPpL/microservices_course_chat-server/grpc/source/
COPY go.sum /github.com/HpPpL/microservices_course_chat-server/grpc/source/
WORKDIR /github.com/HpPpL/microservices_course_chat-server/grpc/source/

RUN go mod download
RUN go build -o ./bin/crud_chat_server ./cmd/grpc_server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/HpPpL/microservices_course_chat-server/grpc/source/bin/crud_chat_server .

CMD ["./crud_chat_server"]