FROM golang:latest as builder
WORKDIR /go/src/github.com/alekssaul/go-header-printer
COPY . .
RUN mkdir -p /app
RUN CGO_ENABLED=0 GOOS=linux go test . 
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/go-header-printer .


FROM alpine:latest
RUN apk update ;  apk add --no-cache ca-certificates ; update-ca-certificates ; mkdir /app
WORKDIR /app
COPY --from=builder /app .
COPY ./go-header-printer /app/go-header-printer
CMD /app/go-header-printer
EXPOSE 8080