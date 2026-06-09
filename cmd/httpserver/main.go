package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bntrtm/http-from-tcp/internal/headers"
	"github.com/bntrtm/http-from-tcp/internal/request"
	"github.com/bntrtm/http-from-tcp/internal/response"
	"github.com/bntrtm/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func proxyHandler(w *response.Writer, r *request.Request) {
	target := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handlerInternalServerError(w, r)
		return
	}
	defer resp.Body.Close()

	_ = w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	h.Remove("Content-Length")
	_ = w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	body := make([]byte, 0)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			body = append(body, buffer[:n]...)
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	sha256 := fmt.Sprintf("%x", sha256.Sum256(body))
	t := headers.NewHeaders()
	t.Override("X-Content-SHA256", sha256)
	t.Override("X-Content-Length", fmt.Sprintf("%d", len(body)))
	err = w.WriteTrailers(t)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
}

func handler(w *response.Writer, r *request.Request) {
	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, r)
		return
	}
	if r.RequestLine.RequestTarget == "/yourproblem" {
		handlerBadRequest(w, r)
		return
	}
	if r.RequestLine.RequestTarget == "/myproblem" {
		handlerInternalServerError(w, r)
		return
	}
	handlerOK(w, r)
}

func handlerBadRequest(w *response.Writer, _ *request.Request) {
	_ = w.WriteStatusLine(response.StatusBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	_ = w.WriteHeaders(h)
	_, _ = w.WriteBody(body)
}

func handlerInternalServerError(w *response.Writer, _ *request.Request) {
	_ = w.WriteStatusLine(response.StatusInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	_ = w.WriteHeaders(h)
	_, _ = w.WriteBody(body)
}

func handlerOK(w *response.Writer, _ *request.Request) {
	_ = w.WriteStatusLine(response.StatusOK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	_ = w.WriteHeaders(h)
	_, _ = w.WriteBody(body)
}
