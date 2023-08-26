FROM debian:12.1

RUN apt update -y && apt install nginx -y

EXPOSE 8089

CMD ["nginx", "-g", "daemon off;"]