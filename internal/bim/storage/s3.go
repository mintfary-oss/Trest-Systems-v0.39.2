package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// S3Store is a small dependency-free AWS Signature V4 client compatible with MinIO and S3.
type S3Store struct {
	Endpoint, Bucket, Region, AccessKey, SecretKey string
	Client                                         *http.Client
}

func (s S3Store) Put(ctx context.Context, key string, r io.Reader) (Object, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Object{}, err
	}
	sum := sha256.Sum256(b)
	hexsum := hex.EncodeToString(sum[:])
	req, err := s.request(ctx, http.MethodPut, key, bytes.NewReader(b), hexsum)
	if err != nil {
		return Object{}, err
	}
	resp, err := (func() (*http.Response, error) {
		c := s.Client
		if c == nil {
			c = http.DefaultClient
		}
		return c.Do(req)
	})()
	if err != nil {
		return Object{}, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Object{}, fmt.Errorf("s3 put: %s", resp.Status)
	}
	return Object{Key: key, Size: int64(len(b)), SHA256: hexsum}, nil
}
func (s S3Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	req, err := s.request(ctx, http.MethodGet, key, nil, "")
	if err != nil {
		return nil, err
	}
	c := s.Client
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("s3 get: %s", resp.Status)
	}
	return resp.Body, nil
}
func (s S3Store) request(ctx context.Context, method, key string, body io.Reader, payloadHash string) (*http.Request, error) {
	if s.Endpoint == "" || s.Bucket == "" || s.AccessKey == "" || s.SecretKey == "" {
		return nil, fmt.Errorf("s3 endpoint, bucket, access key and secret key are required")
	}
	if s.Region == "" {
		s.Region = "us-east-1"
	}
	base, err := url.Parse(strings.TrimRight(s.Endpoint, "/"))
	if err != nil {
		return nil, err
	}
	base.Path = path.Join(base.Path, s.Bucket, key)
	req, err := http.NewRequestWithContext(ctx, method, base.String(), body)
	if err != nil {
		return nil, err
	}
	if payloadHash == "" {
		payloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	now := time.Now().UTC()
	amz := now.Format("20060102T150405Z")
	date := now.Format("20060102")
	req.Header.Set("Host", base.Host)
	req.Header.Set("x-amz-date", amz)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	signed := "host;x-amz-content-sha256;x-amz-date"
	canonical := "" + method + "\n" + base.EscapedPath() + "\n\n" + "host:" + base.Host + "\n" + "x-amz-content-sha256:" + payloadHash + "\n" + "x-amz-date:" + amz + "\n\n" + signed + "\n" + payloadHash
	ch := sha256.Sum256([]byte(canonical))
	scope := date + "/" + s.Region + "/s3/aws4_request"
	sts := "AWS4-HMAC-SHA256\n" + amz + "\n" + scope + "\n" + hex.EncodeToString(ch[:])
	kDate := hmacSHA256([]byte("AWS4"+s.SecretKey), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(s.Region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	sig := hex.EncodeToString(hmacSHA256(kSigning, []byte(sts)))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.AccessKey+"/"+scope+", SignedHeaders="+signed+", Signature="+sig)
	return req, nil
}
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
