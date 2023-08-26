FROM debian:12.1

RUN apt update -y && apt install redis-server ca-certificates -y

EXPOSE 6379

CMD ["redis-server", "/etc/redis/redis.conf"]