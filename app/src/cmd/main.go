package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

func SetKey(ctx context.Context, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can't parse key value", http.StatusBadRequest)
			return
		}

		var jsonData map[string]string
		if err := json.Unmarshal(resp, &jsonData); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			log.Fatal(err)
		}

		for key, value := range jsonData {
			if err = rdb.Set(ctx, key, value, 0).Err(); err != nil {
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
				log.Fatal(err)
			}
			_, _ = io.WriteString(w, fmt.Sprintf("Successfully added %s=%s\n", key, value))
		}
	}
}

func GetKey(ctx context.Context, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			log.Fatal(err)
		}

		_, _ = io.WriteString(w, val)
	}
}

func DelKey(ctx context.Context, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			log.Fatal(err)
		} else if res == 0 {
			http.Error(w, "Key doesn't exists", http.StatusNotFound)
			return
		}

		if err := rdb.Del(ctx, keyData.Key).Err(); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			log.Fatal(err)
		}

		_, _ = io.WriteString(w, "Successfully removed "+keyData.Key)
	}
}
func SetupRedis(ctx context.Context, config *AppConfig) *redis.Client {
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

	address := fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort)
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
	config, err := LoadConfig("config.yaml")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	rdb := SetupRedis(ctx, &config)

	defer rdb.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	})

	http.HandleFunc("/get_key", GetKey(ctx, rdb))
	http.HandleFunc("/set_key", SetKey(ctx, rdb))
	http.HandleFunc("/del_key", DelKey(ctx, rdb))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}
