package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

func main() {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6495ed"))

	urlParam := flag.String("u", "", "url to connect")
	messageParam := flag.String("m", "", "data send to server")
	queryParam := flag.String("q", "", "query params")
	verboseParam := flag.Bool("v", false, "show response server headers")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	urlParse, err := url.Parse(*urlParam)
	if err != nil {
		log.Fatal(err)
	}

	u := url.URL{
		Scheme:   urlParse.Scheme,
		Host:     urlParse.Host,
		Path:     urlParse.Path,
		RawQuery: *queryParam,
	}
	fmt.Printf("connecting to %s\n", u.String())

	header := http.Header{
		"Authorization": {"Bearer YourAccessToken"},
		"CustomHeader":  {"CustomValue"},
	}

	c, response, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	if *verboseParam {
		sh := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff69b4"))

		fmt.Println()
		fmt.Println(sh.Render("RESPONSE HEADERS:"))

		fmt.Println("status -> ", style.Render(response.Status))
		fmt.Println("protocol ->", style.Render(response.Proto))
		fmt.Println("content length -> ", style.Render(string(response.ContentLength)))
		for key, values := range response.Header {
			r := strings.ReplaceAll(key, " ", "")
			fmt.Println(r, "->", style.Render(values...))
		}
	}

	if *messageParam != "" {
		err = c.WriteMessage(websocket.TextMessage, []byte(*messageParam))
		if err != nil {
			log.Fatal("payload send error: ", err)
		}
		fmt.Println("payload send: ", *messageParam)
		return
	}

	done := make(chan struct{})

	fmt.Println()
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("message received: %s\n", message)
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			// Close the connection gracefully before exiting
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
