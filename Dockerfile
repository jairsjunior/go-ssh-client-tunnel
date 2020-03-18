FROM golang:1.13.4 as builder

RUN mkdir /api
WORKDIR /api

ADD go.mod .
ADD go.sum .

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /go/bin/ssh-client .

FROM containous/whoami:v1.5.0 as server


FROM alpine:3.9.5

COPY --from=builder /go/bin/ssh-client /app/

COPY --from=server /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=server /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=server /whoami .

COPY checkIsUp.sh /app/

ENV SSH_SERVER_HOST=ssh-server
ENV SSH_SERVER_PORT=2222
ENV SSH_LOCAL_HOST=localhost
ENV SSH_LOCAL_PORT=8080
ENV SSH_REMOTE_HOST=localhost
ENV SSH_REMOTE_PORT=8080
ENV SSH_USER=convid19
ENV SSH_PASSWORD=c0nv1d19
ENV SSH_MODE=remote

WORKDIR /app
CMD ["sh","checkIsUp.sh"]
