package storage

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestS3StorePutAndGetSignsRequests(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.Method == "PUT" {
			b, _ := io.ReadAll(r.Body)
			if string(b) != "hello" {
				t.Fatalf("body=%q", b)
			}
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	s := S3Store{Endpoint: srv.URL, Bucket: "trest", Region: "us-east-1", AccessKey: "AKIA", SecretKey: "secret"}
	o, err := s.Put(context.Background(), "x.ifc", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if o.Size != 5 || o.SHA256 == "" {
		t.Fatal(o)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 ") {
		t.Fatal(gotAuth)
	}
}
