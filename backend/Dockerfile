FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache gcc musl-dev sqlite-dev
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./main.go

FROM alpine:3.21

WORKDIR /app
RUN apk add --no-cache sqlite-libs ca-certificates
COPY --from=builder /app/server ./server
EXPOSE 8080
CMD ["./server"]
