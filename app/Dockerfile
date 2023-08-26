FROM debian:12.1

RUN apt-get update && apt-get install wget -y

ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.21.0

RUN set -eux; \
	export arch="$(dpkg --print-architecture)"; \
	export url="https://golang.org/dl/go${GOLANG_VERSION}.linux-${arch}.tar.gz"; \
    wget -O go.tgz "$url" --progress=dot:giga; \
    tar -C /usr/local -xzf go.tgz; \
	rm go.tgz;

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

WORKDIR /app

COPY ./src/go-app .

RUN go mod download
RUN go mod tidy
RUN go build -o main .

CMD ["./main"]