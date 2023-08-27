package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type DelKeyData struct {
	Key string `json:"key"`
}

func set_key(rdb *redis.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "400 Bad request", http.StatusBadRequest)
			return
		}

		var data map[string]string
		if err := json.Unmarshal(bytes, &data); err != nil {
			panic(err)
		}

		response_str := ""
		for key, value := range data {
			err = rdb.Set(ctx, key, value, 0).Err()
			if err != nil {
				panic(err)
			}
			response_str += fmt.Sprintf("Successfully added %s=%s\n", key, value)
		}

		io.WriteString(w, response_str)
	}
}

func get_key(rdb *redis.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		query_key := r.URL.Query().Get("key")
		if query_key == "" {
			http.Error(w, "400 Bad request", http.StatusBadRequest)
			return
		}

		val, err := rdb.Get(ctx, query_key).Result()
		if err == redis.Nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		} else if err != nil {
			panic(err)
		} else {
			io.WriteString(w, val)
		}
	}
}

func del_key(rdb *redis.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var keyData DelKeyData

		json_decoder := json.NewDecoder(r.Body)

		err := json_decoder.Decode(&keyData)
		if err != nil || keyData.Key == "" {
			http.Error(w, "400 Bad request", http.StatusBadRequest)
			return
		}

		res, err := rdb.Exists(ctx, keyData.Key).Result()
		if res == 0 {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			panic(err)
		}

		err = rdb.Del(ctx, keyData.Key).Err()
		if err != nil {
			panic(err)
		}

		io.WriteString(w, fmt.Sprintf("Successfully removed %s", keyData.Key))
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair("tls/redis.crt", "tls/redis.key")
	if err != nil {
		panic(err)
	}

	caCert, err := os.ReadFile("tls/ca.crt")
	if err != nil {
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:      "redis:6379",
		Username:  os.Getenv("REDIS_USER"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		DB:        0,
		TLSConfig: tlsConfig,
	})

	defer rdb.Close()

	ctx := context.Background()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	})

	http.HandleFunc("/get_key", get_key(rdb, ctx))
	http.HandleFunc("/set_key", set_key(rdb, ctx))
	http.HandleFunc("/del_key", del_key(rdb, ctx))

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
