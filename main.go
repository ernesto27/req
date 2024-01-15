package main

import (
	"flag"
	"fmt"
)

type Params struct {
	typeP   string
	url     string
	query   string
	header  string
	message string
}

func getProtocol(p Params) Protocol {
	switch p.typeP {
	case "ws":
		return NewWebsocket(p)
	case "gq":
		return NewGrapQL(p)
	}
	return nil
}

func main() {
	typeParam := flag.String("t", "", "type connection (ws, graphql)")
	urlParam := flag.String("u", "", "url to connect")
	messageParam := flag.String("m", "", "data send to server")
	queryParam := flag.String("q", "", "query params")
	verboseParam := flag.Bool("v", false, "show response server headers")
	headerP := flag.String("h", "", "header params")
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
	}

	p := getProtocol(params)
	defer p.Close()

	resp, err := p.RequestResponse()
	if err != nil {
		panic(err)
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
