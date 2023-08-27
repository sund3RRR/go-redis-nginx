FROM debian:12.1

RUN apt-get update -y && apt-get install redis-server -y

EXPOSE 6379

CMD ["redis-server", "/etc/redis/redis.conf"]