package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Params struct {
	typeP   string
	url     string
	query   string
	header  string
	message string
	method  string
}

func getProtocol(p Params) Protocol {
	switch p.typeP {
	case "ws":
		return NewWebsocket(p)
	case "gq":
		return NewGrapQL(p)
	case "http":
		return NewHTTP(p)
	}
	return nil
}

func main() {
	typeParam := flag.String("t", "http", "type connection (ws, gq, http)")
	urlParam := flag.String("u", "", "url to connect")
	messageParam := flag.String("p", "", "data send to server")
	queryParam := flag.String("q", "", "query params")
	verboseParam := flag.Bool("v", false, "show response server headers")
	headerP := flag.String("h", "", "header params")
	method := flag.String("m", "GET", "method request")

	flag.Parse()

	if *typeParam == "" {
		fmt.Println("type connection is required - ws, gq")
		return
	}

	params := Params{
		typeP:   *typeParam,
		url:     *urlParam,
		query:   *queryParam,
		header:  *headerP,
		message: *messageParam,
		method:  *method,
	}

	p := getProtocol(params)
	defer p.Close()

	resp, err := p.RequestResponse()
	if err != nil {
		var styleErr = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#e05074")).Padding(1).MarginBottom(1).MarginTop(1)

		fmt.Println(styleErr.Render(err.Error()))
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(resp)

	if *messageParam == "" {
		p.OnMessageReceived()
	}

	if *verboseParam {
		p.PrintHeaderResponse()

	}

}
