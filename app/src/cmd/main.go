package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func HasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

func setKeyHandler(ctx context.Context, rdb *redis.Client, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request", zap.String("URL", r.URL.Path), zap.String("Method", r.Method))

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !HasContentType(r, "application/json") {
			http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
			return
		}

		resp, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't parse key value", http.StatusBadRequest)
			return
		}

		var jsonData map[string]string
		if err := json.Unmarshal(resp, &jsonData); err != nil {
			http.Error(w, "Can't parse key value", http.StatusBadRequest)
			return
		}

		for key, value := range jsonData {
			if err = rdb.Set(ctx, key, value, 0).Err(); err != nil {
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
				logger.Error("An error occured while executing redis.Set", zap.Error(err))
				return
			}
			_, _ = io.WriteString(w, fmt.Sprintf("Successfully added %s=%s\n", key, value))
		}
	}
}

func getKeyHandler(ctx context.Context, rdb *redis.Client, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request", zap.String("URL", r.URL.Path), zap.String("Method", r.Method))

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		queryKey := r.URL.Query().Get("key")
		if queryKey == "" {
			http.Error(w, "Can't parse key", http.StatusBadRequest)
			return
		}

		val, err := rdb.Get(ctx, queryKey).Result()
		if err != nil {
			if err == redis.Nil {
				http.Error(w, "Key not found", http.StatusNotFound)
				return
			}
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			logger.Error("An error occured while executing redis.Get", zap.Error(err))
			return
		}

		_, _ = io.WriteString(w, val)
	}
}

func delKeyHandler(ctx context.Context, rdb *redis.Client, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request", zap.String("URL", r.URL.Path), zap.String("Method", r.Method))

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !HasContentType(r, "application/json") {
			http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
			return
		}

		type DelKeyData struct {
			Key string `json:"key"`
		}

		var keyData DelKeyData

		jsonDecoder := json.NewDecoder(r.Body)

		if err := jsonDecoder.Decode(&keyData); err != nil || keyData.Key == "" {
			http.Error(w, "Can't parse key", http.StatusBadRequest)
			return
		}

		res, err := rdb.Exists(ctx, keyData.Key).Result()
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			logger.Error("An error occured while executing redis.Exists", zap.Error(err))
			return
		} else if res == 0 {
			http.Error(w, "Key doesn't exists", http.StatusNotFound)
			return
		}

		if err := rdb.Del(ctx, keyData.Key).Err(); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			logger.Error("An error occured while executing redis.Del", zap.Error(err))
			return
		}

		_, _ = io.WriteString(w, "Successfully removed "+keyData.Key)
	}
}

func createTLSConfig(config *AppConfig) *tls.Config {
	if !config.UseRedisTLS {
		return nil
	}

	cert, err := tls.LoadX509KeyPair("tls/redis.crt", "tls/redis.key")
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := os.ReadFile("tls/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return tlsConfig
}

func setupRedis(ctx context.Context, config *AppConfig) *redis.Client {
	address := fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort)

	tlsConfig := createTLSConfig(config)

	rdb := redis.NewClient(&redis.Options{
		Addr:      address,
		Username:  config.RedisUser,
		Password:  config.RedisPassword,
		TLSConfig: tlsConfig,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal(err)
	}

	return rdb
}

func main() {
	config, err := NewConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	logger, err := config.ZapConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	ctx := context.Background()

	rdb := setupRedis(ctx, config)
	defer rdb.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	})

	http.HandleFunc("/get_key", getKeyHandler(ctx, rdb, logger))
	http.HandleFunc("/set_key", setKeyHandler(ctx, rdb, logger))
	http.HandleFunc("/del_key", delKeyHandler(ctx, rdb, logger))

	logger.Info("Starting server", zap.String("Server port", fmt.Sprintf("%d", config.ServerPort)))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil); err != nil {
		logger.Error("An error occured while starting a server", zap.Error(err))
		os.Exit(1)
	}
}
