package main

import (
	"net/http"
	"testing"
)

// Can test with go test -v ./cmd/api
// This is unfinished...ran out of time and can just test with Swagger UI
func TestGetCountryByIPHandler(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount() //lets us copy our routes for easy testing...the joy of modularity and decoupling
	t.Run("should not allow empty ip", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/"+appVersion+"/ip-map/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected the response code to be %d but got %d", http.StatusNotFound, rr.Code)
		}

	})
}
