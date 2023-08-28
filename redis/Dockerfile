FROM debian:12.1

RUN set -eux; \
    apt-get update -y; \
    apt-get install redis-server -y;

EXPOSE 6379

USER redis

CMD ["redis-server", "/etc/redis/redis.conf"]