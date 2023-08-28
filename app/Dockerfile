FROM debian:12.1

RUN apt-get update && apt-get install wget -y

ENV GOLANG_VERSION 1.21.0

ENV TZ=Europe/Moscow

RUN set -eux; \
	export arch="$(dpkg --print-architecture)"; \
	export url="https://golang.org/dl/go${GOLANG_VERSION}.linux-${arch}.tar.gz"; \
	wget -O go.tgz "$url" --progress=dot:giga; \
    tar -C /usr/local -xzf go.tgz; \
	rm go.tgz;

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$GOPATH/bin:$PATH

WORKDIR /app

COPY ./src/cmd/config.go .
COPY ./src/cmd/main.go .
COPY ./src/go.mod .

RUN set -eux; \
	go mod download; \
	go mod tidy; \
	go build -o main .;

CMD ["./main"]