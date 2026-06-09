package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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

func handler(w *response.Writer, r *request.Request) {
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
