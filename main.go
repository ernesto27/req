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

var styleInfo = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Bold(true).
	Background(lipgloss.Color("#6699FF")).Padding(1).MarginBottom(1).MarginTop(1).MarginLeft(1)

var styleErr = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FAFAFA")).
	Bold(true).
	Background(lipgloss.Color("#e05074")).Padding(1).MarginBottom(1).MarginTop(1).MarginLeft(1)

var styleMessage = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#606683")).Padding(1).MarginBottom(1).MarginTop(1).MarginLeft(1)

func main() {
	typeParam := flag.String("t", "http", "type connection (ws, gq, http)")
	urlParam := flag.String("u", "", "url to connect")
	messageParam := flag.String("p", "", "data send to server")
	queryParam := flag.String("q", "", "query params")
	verboseParam := flag.Bool("v", false, "show response server headers")
	headerP := flag.String("h", "", "header params")
	method := flag.String("m", "GET", "method request")

	flag.Parse()

	validProtocos := []string{"ws", "gq", "http"}

	valid := false
	for _, v := range validProtocos {
		if *typeParam == v {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Println(styleErr.Render("Invalid protocol, valid are: http, ws, gq"))
		os.Exit(0)
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
		fmt.Println(styleErr.Render(err.Error()))
		os.Exit(1)
	}

	fmt.Println(resp)

	if *messageParam == "" {
		p.OnMessageReceived()
	}

	if *verboseParam {
		p.PrintHeaderResponse()
	}

}
