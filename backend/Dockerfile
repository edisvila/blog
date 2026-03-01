FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o blog_server

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/blog_server .
EXPOSE 8080
CMD ["./blog_server"]