FROM golang:1.18.3-alpine3.16

WORKDIR /app
COPY pgssl .

RUN go build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pgssl .

FROM alpine:3.16.0
RUN apk --no-cache add ca-certificates
WORKDIR /pgssl
COPY --from=0 /app .

EXPOSE 15432

ENV PG_HOST localhost
ENV PG_PORT 5432

CMD ./pgssl -p ${PG_HOST}:${PG_PORT} -l :5432 -c ca-certificate.crt
