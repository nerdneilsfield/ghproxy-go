FROM alpine:latest

COPY ghproxy-go /usr/bin/ghproxy-go

CMD ["/usr/bin/ghproxy-go", "run"]
