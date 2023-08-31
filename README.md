# go-redis-nginx

*Практическое задание на позицию стажёра в IT-Security: DevOps в Avito Tech* 

## О задании
### Задача
Разработать минималистичное приложение-интерфейс для работы с `Redis`'ом с проксированием трафика. Выполнить его сборку и развёртывание.

### Требования
✅ В контейнерах, в качестве базового образа, используйте официальный образ `Debian`

✅ Развертывание и сборка должны выполняться по средствам `Docker Compose` (compose file version >= `3.3`)

✅ Приложение реализовано на `Golang`

✅ В `Redis`'e работаем только со строками

### Опциональные требования
✅ На `Redis`'e поддержана аутентификация

✅ `Redis` и приложение "общаются" по зашифрованному каналу
(`TLS`-соединение)

## О проекте
### Версии
- Версия `compose` файла - `3.8`
- Версия образа `Debian` - `12.1`
- Версия `Golang` - `1.21.0`
- Версия `nginx` - `1.22.1`
- Версия `redis-server` - `7.0.11`

### Redis
Ключ и сертификаты для `TLS` шифрования были сгенерированы командами: 
```bash
openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout redis.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=redis" -out server.csr

openssl x509 -req -extfile <(printf "subjectAltName=DNS:redis") -days 365 -in redis.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out redis.crt
```

Подключение по `TLS` опционально, его можно отключить в `config.yaml` для приложения `go` и конфиге самого `Redis`'a

#### **`app/src/cmd/config.yaml`**
```yaml
use_redis_tls: false
```

#### **`app/src/redis-server/redis.conf`**
```
port 6379
tls-port 0
```

Настройки пользователей `Redis`'a находятся в файле `app/src/redis-server/acl.conf`

Использование аутентификации также опционально и при необходимости можно оставить поля аутентификации пустыми в конфиге `go` приложения:

#### **`app/src/cmd/config.yaml`**
```yaml
redis_user: ""
redis_password: ""
```

### Приложение Go
Были использованы сторонние библиотеки:
- [go-redis](https://github.com/redis/go-redis) для операций с `Redis`
- [go-yaml](https://github.com/go-yaml/yaml) для десереализации `.yaml` файла, в котором находится конфиг приложения
- [zap](https://github.com/uber-go/zap) для логирования

`zap` настроен в *Development* режим:
```golang
zapConfig := zap.NewDevelopmentConfig()
zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
```

## Установка
Сборка и запуск проекта были протестированы на архитектурах `amd64` и `arm64`

### Установка зависимостей
**`.deb` based дистрибутивы**:
```bash
sudo apt update && sudo apt install git docker docker-compose -y
```
**MacOS Intel/Apple Silicon**:

Скачайте и установите `Docker Desktop` с [официального](https://www.docker.com/products/docker-desktop) сайта

### Клонирование репозитория
```bash
git clone https://github.com/sund3RRR/go-redis-nginx.git
cd go-redis-nginx
```

### Сборка проекта
```bash
docker-compose build
```

### Запуск
```bash
docker-compose up
```
