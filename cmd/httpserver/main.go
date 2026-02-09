package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(req *request.Request, w *response.Writer) error {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return handleYourProblem(w)
	case "/myproblem":
		return handleMyProblem(w)
	default:
		return handleSuccess(w)
	}
}

func handleYourProblem(w *response.Writer) error {
	html := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`
	body := []byte(html)

	err := w.WriteStatusLine(response.StatusBadRequest)
	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("content-type", "text/html")

	err = w.WriteHeaders(headers)
	if err != nil {
		return err
	}

	_, err = w.WriteBody(body)
	return err
}

func handleMyProblem(w *response.Writer) error {
	html := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`
	body := []byte(html)

	err := w.WriteStatusLine(response.StatusInternalServerError)
	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("content-type", "text/html")

	err = w.WriteHeaders(headers)
	if err != nil {
		return err
	}

	_, err = w.WriteBody(body)
	return err
}

func handleSuccess(w *response.Writer) error {
	html := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`
	body := []byte(html)

	err := w.WriteStatusLine(response.StatusOK)
	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("content-type", "text/html")

	err = w.WriteHeaders(headers)
	if err != nil {
		return err
	}

	_, err = w.WriteBody(body)
	return err
}
