FROM debian:12.1

RUN set -eux; \
    apt-get update -y; \
    apt-get install nginx -y;

EXPOSE 8089

CMD ["nginx", "-g", "daemon off;", "-c", "/etc/nginx/nginx.conf"]