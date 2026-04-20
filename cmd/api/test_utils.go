package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delgme1vzw/ip-origin-validation/internal/datastore"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	testDatastore := datastore.NewDummyStorage()

	return &application{
		store: testDatastore,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}
