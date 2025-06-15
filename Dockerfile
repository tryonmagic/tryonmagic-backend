FROM golang:1.23 AS build

WORKDIR /app
COPY . .
RUN go clean --modcache
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build src/main.go

FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /app
COPY --from=build /app/main .

EXPOSE 8000
CMD ["./main"]