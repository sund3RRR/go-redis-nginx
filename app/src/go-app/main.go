package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type DelKeyData struct {
	Key string `json:"key"`
}

func set_key(rdb *redis.Client) http.HandlerFunc {
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
			response_str += fmt.Sprintf("Successfully added %s=%s", key, value)
		}

		io.WriteString(w, response_str)
	}
}

func get_key(rdb *redis.Client) http.HandlerFunc {
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

func del_key(rdb *redis.Client) http.HandlerFunc {
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
		err = rdb.Exists(ctx, keyData.Key).Err()
		if err != nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}
		err = rdb.Del(ctx, keyData.Key).Err()
		if err != nil {
			panic(err)
		}

		io.WriteString(w, fmt.Sprintf("Successfully removed %s", keyData.Key))
	}
}

func main() {
	// cert, err := tls.LoadX509KeyPair("tls/redis.crt", "tls/redis.key")
	// if err != nil {
	// 	panic(err)
	// }

	// // Load CA cert
	// caCert, err := os.ReadFile("tls/ca.crt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		// MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		// Certificates: []tls.Certificate{cert},
		// RootCAs:      caCertPool,
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:      "redis:6378",
		Password:  "",
		DB:        0,
		TLSConfig: tlsConfig,
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	})

	http.HandleFunc("/get_key", get_key(rdb))
	http.HandleFunc("/set_key", set_key(rdb))
	http.HandleFunc("/del_key", del_key(rdb))

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pong)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
