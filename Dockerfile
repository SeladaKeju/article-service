FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /article-service ./cmd

FROM alpine:3.21

RUN adduser -D -H -u 10001 app
COPY --from=build /article-service /article-service
USER app
EXPOSE 8080
ENTRYPOINT ["/article-service"]
