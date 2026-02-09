package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer f.Close()
		defer close(out)

		buffer := make([]byte, 8)
		currentLine := ""

		for {
			n, err := f.Read(buffer)

			if n > 0 {
				chunk := string(buffer[:n])
				parts := strings.Split(chunk, "\n")

				for i := 0; i < len(parts)-1; i++ {
					currentLine += parts[i]
					out <- currentLine
					currentLine = ""
				}

				currentLine += parts[len(parts)-1]
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				return
			}
		}

		if currentLine != "" {
			out <- currentLine
		}
	}()

	return out
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}

		fmt.Println("Connection accepted")

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}

		fmt.Println("Connection closed")
	}
}
