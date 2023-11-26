package htmldownloader

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTMLDownloader_DownloadSuccess(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a sample HTML response
		html := "<html><body><h1>Hello, World!</h1></body></html>"
		w.Write([]byte(html))
	}))
	defer server.Close()

	// Create a new HTMLDownloader instance
	downloader := &HTMLDownloader{
		url: server.URL,
	}

	// Call the Download method
	html, err := downloader.Download()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify the downloaded HTML content
	expectedHTML := "<html><body><h1>Hello, World!</h1></body></html>"
	if html != expectedHTML {
		t.Errorf("Unexpected HTML content. Expected: %s, Got: %s", expectedHTML, html)
	}
}

func TestHTMLDownloader_DownloadError(t *testing.T) {
	// Create a new HTMLDownloader instance
	downloader := &HTMLDownloader{
		url: "http://noexistssite77777",
	}

	// Call the Download method
	_, err := downloader.Download()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
