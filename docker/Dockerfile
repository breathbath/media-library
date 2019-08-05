FROM golang:1.12
WORKDIR /go/mediaService
COPY ../. .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o media .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/mediaService/media .