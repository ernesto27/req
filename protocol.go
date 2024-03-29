package main

import (
	"bytes"
	"context"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fullstorydev/grpcurl"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type Protocol interface {
	RequestResponse() (string, error)
	OnMessageReceived()
	PrintHeaderResponse()
	Close()
	Download() error
}

type GraphQL struct {
	url      string
	query    string
	httpResp *http.Response
	response string
	timeout  int
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

	resp, httpResp, err := DoRequest("POST", string(payloadBytes), g.url, "", "", "", g.timeout)
	g.httpResp = httpResp
	return resp, err
}

func (g *GraphQL) OnMessageReceived() {}

func (g *GraphQL) Download() error {
	err := saveToFile(g.response, g.url)
	if err != nil {
		return err
	}

	return nil
}

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
	fmt.Print(styleInfo.Render("connecting to ", u.String()))

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

	if params.userAgent != "" {
		header.Add("User-Agent", params.userAgent)
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
	return "", nil
}

func (w *Websocket) OnMessageReceived() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := w.client.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Print(styleMessage.Render("message received:", string(message)))
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

func (w *Websocket) Download() error { return nil }

func (w *Websocket) Close() {
	w.client.Close()
}

type HTTP struct {
	url       string
	method    string
	httpResp  *http.Response
	headers   string
	body      string
	file      string
	response  string
	userAgent string
	timeout   int
}

func NewHTTP(params Params) *HTTP {
	return &HTTP{
		url:       params.url,
		method:    params.method,
		headers:   params.header,
		body:      params.message,
		file:      params.file,
		userAgent: params.userAgent,
		timeout:   params.timeout,
	}
}

func (h *HTTP) RequestResponse() (string, error) {
	resp, httpResp, err := DoRequest(h.method, h.body, h.url, h.userAgent, h.headers, h.file, h.timeout)
	h.httpResp = httpResp

	return resp, err
}

func (h *HTTP) OnMessageReceived() {}

func (h *HTTP) Download() error {
	err := saveToFile(h.response, h.url)
	if err != nil {
		return err
	}

	return nil
}

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
	fmt.Println(sh.Render("RESPONSE SERVER:"))
	fmt.Println("status -> ", style.Render(r.Status))
	fmt.Println("protocol ->", style.Render(r.Proto))
	fmt.Println("content length -> ", style.Render(fmt.Sprint(r.ContentLength)))
	for key, values := range r.Header {
		r := strings.ReplaceAll(key, " ", "")
		fmt.Println(r, "->", style.Render(values...))
	}
	fmt.Println()
}

type GRPC struct {
	cc         *grpc.ClientConn
	url        string
	importPath string
	proto      string
	message    string
	methodName string
	verbose    bool
	response   string
}

func NewGRPC(params Params) *GRPC {
	return &GRPC{
		url:        params.url,
		importPath: params.importPath,
		proto:      params.proto,
		message:    params.message,
		methodName: params.methodName,
		verbose:    params.verbose,
	}
}

func (g *GRPC) RequestResponse() (string, error) {
	ctx := context.Background()
	dial := func() (*grpc.ClientConn, error) {
		dialTime := 10 * time.Second
		ctx, cancel := context.WithTimeout(ctx, dialTime)
		defer cancel()
		var opts []grpc.DialOption

		network := "tcp"
		var creds credentials.TransportCredentials

		grpcurlUA := "grpcurl/" + "dev build <no version set>"

		opts = append(opts, grpc.WithUserAgent(grpcurlUA))
		target := g.url
		cc, err := grpcurl.BlockingDial(ctx, network, target, creds, opts...)
		if err != nil {
			return nil, err
		}
		return cc, nil
	}

	var descSource grpcurl.DescriptorSource
	var fileSource grpcurl.DescriptorSource

	importPaths := []string{g.importPath}
	protoFiles := []string{g.proto}

	var err error
	fileSource, err = grpcurl.DescriptorSourceFromProtoFiles(importPaths, protoFiles...)
	if err != nil {
		return "", err
	}

	descSource = fileSource

	if g.cc == nil {
		g.cc, err = dial()
		if err != nil {
			return "", err
		}
	}

	in := strings.NewReader(g.message)

	options := grpcurl.FormatOptions{
		EmitJSONDefaultFields: true,
		IncludeTextSeparator:  true,
		AllowUnknownFields:    true,
	}
	rf, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format("json"), descSource, in, options)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	v := 0
	if g.verbose {
		v = 1
	}

	h := &grpcurl.DefaultEventHandler{
		Out:            &buf,
		Formatter:      formatter,
		VerbosityLevel: v,
	}

	symbol := g.methodName
	err = grpcurl.InvokeRPC(ctx, descSource, g.cc, symbol, append([]string{}, []string{}...), h, rf.Next)

	if err != nil {
		if errStatus, ok := status.FromError(err); ok && false {
			h.Status = errStatus
		} else {
			return "", err
		}
	}

	if h.Status.Code() != codes.OK {
		grpcurl.PrintStatus(os.Stderr, h.Status, formatter)
	}

	g.response = buf.String()

	return g.response, nil
}

func (g *GRPC) OnMessageReceived()   {}
func (g *GRPC) PrintHeaderResponse() {}
func (g *GRPC) Download() error {
	err := saveToFile(g.response, g.url)
	if err != nil {
		return err
	}

	return nil
}

func (g *GRPC) Close() {
	if g.cc != nil {
		g.cc.Close()
		g.cc = nil
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

func GetRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func saveToFile(data string, urlVal string) error {
	parsedURL, err := url.Parse(urlVal)
	if err != nil {
		return err
	}

	path := parsedURL.Path
	hasExtension, name := hasFileExtension(path)

	var fileName string
	if hasExtension {
		fileName = name
	} else {
		fileName = GetRandomString(10)
	}

	err = os.WriteFile(fileName, []byte(data), 0644)
	if err != nil {
		return err
	}

	return nil
}

func hasFileExtension(path string) (bool, string) {
	segments := strings.Split(path, "/")

	if len(segments) == 0 {
		return false, ""
	}

	lastSegment := segments[len(segments)-1]

	parts := strings.Split(lastSegment, ".")
	return len(parts) >= 2, lastSegment
}

func DoRequest(method string, body string, urlV string, userAgent string, headers string, file string, timeout int) (string, *http.Response, error) {
	client := &http.Client{}

	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	m := strings.ToUpper(method)

	validMethods := []string{"GET", "POST", "DELETE", "PUT"}

	found := false
	for _, vm := range validMethods {
		if vm == m {
			found = true
		}
	}

	if !found {
		return "", nil, fmt.Errorf("invalid method %s ", m)
	}

	var js map[string]interface{}
	err := json.Unmarshal([]byte(body), &js)
	var d io.Reader
	var contentType string

	if err != nil {
		items := getHeaders(body)
		formData := url.Values{}
		for _, h := range items {
			formData.Add(h.Key, h.Value)
		}

		d = strings.NewReader(formData.Encode())

		if file != "" {
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)

			pathImg := strings.Split(file, "=")
			if len(pathImg) != 2 {
				return "", nil, fmt.Errorf("invalid file %s, valid file=myfile.png", file)
			}

			file, err := os.Open(pathImg[1])
			if err != nil {
				return "", nil, err
			}
			defer file.Close()

			fileWriter, err := bodyWriter.CreateFormFile(pathImg[0], GetRandomString(10)+".png")
			if err != nil {
				return "", nil, err
			}

			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return "", nil, err
			}

			contentType = bodyWriter.FormDataContentType()
			_ = bodyWriter.Close()
			d = bodyBuf

		} else {
			contentType = "application/x-www-form-urlencoded"
		}

	} else {
		d = bytes.NewBuffer([]byte(body))
		contentType = "application/json"
	}

	req, err := http.NewRequest(m, urlV, d)

	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Content-Type", contentType)
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	headersP := getHeaders(headers)
	for _, h := range headersP {
		req.Header.Add(h.Key, h.Value)
	}

	response, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer response.Body.Close()

	httpResp := response

	bodyResp, err := io.ReadAll(response.Body)
	if err != nil {
		return "", nil, err
	}

	return string(bodyResp), httpResp, nil
}
