package telegram

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUploadPhotoRetryReplaysMultipartBody(t *testing.T) {
	attempts := 0
	bodySizes := make([]int, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		bodySizes = append(bodySizes, len(body))
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts == 1 {
			_, _ = w.Write([]byte(`{"ok":false,"error_code":500,"description":"temporary failure"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":42,"photo":[{"file_id":"file-123","file_unique_id":"unique","width":10,"height":10}]}}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.APIBaseURL = server.URL
	client.Timeout = time.Second
	client.Retry = 1

	fileID, messageID, err := client.UploadPhotoByBytes("-100123", []byte("image-content"), "image.png", "test")
	if err != nil {
		t.Fatalf("UploadPhotoByBytes returned error: %v", err)
	}
	if fileID != "file-123" || messageID != 42 {
		t.Fatalf("unexpected result: fileID=%q messageID=%d", fileID, messageID)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if bodySizes[0] == 0 || bodySizes[1] != bodySizes[0] {
		t.Fatalf("retry body was not replayed: sizes=%v", bodySizes)
	}
}
