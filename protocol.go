package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type Protocol interface {
	RequestResponse() (string, error)
	OnMessageReceived()
	PrintHeaderResponse()
	Close()
}

type GraphQL struct {
	url      string
	query    string
	httpResp *http.Response
}

func NewGrapQL(params Params) *GraphQL {
	return &GraphQL{
		url:   params.url,
		query: params.message,
	}
}

func (g *GraphQL) RequestResponse() (string, error) {
	requestPayload := map[string]interface{}{
		"query": g.query,
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(g.url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	g.httpResp = resp

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (g *GraphQL) OnMessageReceived() {}
func (g *GraphQL) PrintHeaderResponse() {
	printHttpResponse(g.httpResp)
}
func (g *GraphQL) Close() {}

type Websocket struct {
	url          string
	queryParams  string
	headerParams string
	message      string
	client       *websocket.Conn
	httpResp     *http.Response
}

func NewWebsocket(params Params) *Websocket {
	urlParse, err := url.Parse(params.url)
	if err != nil {
		log.Fatal(err)
	}

	u := url.URL{
		Scheme:   urlParse.Scheme,
		Host:     urlParse.Host,
		Path:     urlParse.Path,
		RawQuery: params.query,
	}
	fmt.Printf("connecting to %s\n", u.String())

	header := http.Header{}
	if params.header != "" {
		headers := strings.Split(params.header, "&")
		for _, h := range headers {
			h := strings.Split(h, "=")
			if len(h) != 2 {
				continue
			}
			header.Add(h[0], h[1])
		}
	}

	c, response, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return &Websocket{
		url:          params.url,
		queryParams:  params.query,
		headerParams: params.header,
		client:       c,
		httpResp:     response,
		message:      params.message,
	}
}

func (w *Websocket) RequestResponse() (string, error) {
	err := w.client.WriteMessage(websocket.TextMessage, []byte(w.message))
	if err != nil {
		return "", err
	}
	return w.message, nil
}

func (w *Websocket) OnMessageReceived() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	fmt.Println()
	go func() {
		defer close(done)
		for {
			_, message, err := w.client.ReadMessage()
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
			err := w.client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

func (w *Websocket) PrintHeaderResponse() {
	printHttpResponse(w.httpResp)
}

func (w *Websocket) Close() {
	w.client.Close()
}

type HTTP struct {
	url      string
	method   string
	httpResp *http.Response
	headers  string
	body     string
}

func NewHTTP(params Params) *HTTP {
	return &HTTP{
		url:     params.url,
		method:  params.method,
		headers: params.header,
		body:    params.message,
	}
}

func (h *HTTP) RequestResponse() (string, error) {
	client := &http.Client{}

	m := strings.ToUpper(h.method)

	validMethods := []string{"GET", "POST"}

	found := false
	for _, vm := range validMethods {
		if vm == m {
			found = true
		}
	}

	if !found {
		return "", fmt.Errorf("invalid method %s ", m)
	}

	var js map[string]interface{}
	err := json.Unmarshal([]byte(h.body), &js)
	var d io.Reader
	var contentType string

	if err != nil {
		items := getHeaders(h.body)
		formData := url.Values{}
		for _, h := range items {
			formData.Add(h.Key, h.Value)
		}

		d = strings.NewReader(formData.Encode())
		contentType = "application/x-www-form-urlencoded"
	} else {
		d = bytes.NewBuffer([]byte(h.body))
		contentType = "application/json"
	}

	req, err := http.NewRequest(m, h.url, d)

	req.Header.Set("Content-Type", contentType)
	if err != nil {
		return "", err
	}

	headersP := getHeaders(h.headers)
	for _, h := range headersP {
		req.Header.Add(h.Key, h.Value)
	}

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	h.httpResp = response

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (h *HTTP) OnMessageReceived() {}
func (h *HTTP) PrintHeaderResponse() {
	printHttpResponse(h.httpResp)
}
func (h *HTTP) Close() {}

func printHttpResponse(r *http.Response) {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#6495ed"))

	sh := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ff69b4"))

	fmt.Println()
	fmt.Println(sh.Render("RESPONSE HEADERS:"))
	fmt.Println("status -> ", style.Render(r.Status))
	fmt.Println("protocol ->", style.Render(r.Proto))
	fmt.Println("content length -> ", style.Render(fmt.Sprint(r.ContentLength)))
	for key, values := range r.Header {
		r := strings.ReplaceAll(key, " ", "")
		fmt.Println(r, "->", style.Render(values...))
	}
}

type HeaderP struct {
	Key   string
	Value string
}

func getHeaders(headers string) []HeaderP {
	resp := []HeaderP{}
	if headers != "" {
		items := strings.Split(headers, "&")
		for _, h := range items {
			h := strings.Split(h, "=")
			if len(h) != 2 {
				continue
			}
			resp = append(resp, HeaderP{
				Key:   h[0],
				Value: h[1],
			})
		}
	}

	return resp
}
