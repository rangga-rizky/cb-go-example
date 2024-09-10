package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rangga-rizky/go-cb/util"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	var ctx = context.Background()

	cb := util.NewCB(*rdb)
	cb.Register(ctx, "xixix", 5)

	http.HandleFunc("/", simpleHandler(cb))
	http.ListenAndServe(":8080", nil)
}

func simpleHandler(cb util.CB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ignore favicon.ico requests
		if r.URL.Path == "/favicon.ico" {
			http.NotFound(w, r)
			return
		}

		responseText := "Hello word"
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		_, err := client.Get("http://localhost:9090/")
		if err != nil {
			cb.Count(r.Context(), "xixix")
			counter := cb.GetCounter(r.Context(), "xixix")
			responseText = "your cb error counter: " + strconv.Itoa(counter)
		}

		if cb.IsOpen(r.Context(), "xixix") {
			responseText = "CB opened!"
		}

		fmt.Fprint(w, responseText)
	}
}
