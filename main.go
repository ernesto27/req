package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Params struct {
	typeP      string
	url        string
	query      string
	header     string
	message    string
	method     string
	file       string
	importPath string
	proto      string
	methodName string
	verbose    bool
}

func getProtocol(p Params) Protocol {
	switch p.typeP {
	case "ws":
		return NewWebsocket(p)
	case "gq":
		return NewGrapQL(p)
	case "http":
		return NewHTTP(p)
	case "grpc":
		return NewGRPC(p)
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
	typeParam := flag.String("t", "http", "type connection (ws, gq, http, grpc)")
	urlParam := flag.String("u", "", "url to connect")
	messageParam := flag.String("p", "", "data send to server")
	queryParam := flag.String("q", "", "query params")
	verboseParam := flag.Bool("v", false, "show response server headers")
	headerP := flag.String("h", "", "header params")
	method := flag.String("m", "GET", "method request")
	file := flag.String("f", "", "file path")

	// GRPC exclusive flags
	importPath := flag.String("import-path", "", "The path to a directory from which proto sources can be imported, for use with -proto flags")
	proto := flag.String("proto", "", "The proto file")
	methodName := flag.String("method", "", "The service method to call")

	flag.Parse()

	if *urlParam == "" {
		flag.PrintDefaults()
		return
	}

	validProtocos := []string{"ws", "gq", "http", "grpc"}

	valid := false
	for _, v := range validProtocos {
		if *typeParam == v {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Println(styleErr.Render("Invalid protocol, valid are: http, ws, gq"))
		return
	}

	params := Params{
		typeP:      *typeParam,
		url:        *urlParam,
		query:      *queryParam,
		header:     *headerP,
		message:    *messageParam,
		method:     *method,
		file:       *file,
		importPath: *importPath,
		proto:      *proto,
		methodName: *methodName,
		verbose:    *verboseParam,
	}

	p := getProtocol(params)
	defer p.Close()

	if *messageParam == "" {
		p.OnMessageReceived()
	}

	resp, err := p.RequestResponse()
	if err != nil {
		fmt.Println(styleErr.Render(err.Error()))
		return
	}

	myRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(140),
	)

	if err != nil {
		fmt.Println(styleErr.Render(err.Error()))
		return
	}

	r, err := PrettyJSON(resp)
	if err != nil {
		dr, err := myRender.Render("```html\n" + resp + "\n```\n")
		if err != nil {
			fmt.Println(styleErr.Render(err.Error()))
			return
		}
		fmt.Print(dr)
		if *verboseParam {
			p.PrintHeaderResponse()
		}
		return
	}

	out, err := myRender.Render(r)
	if err != nil {
		fmt.Println(styleErr.Render(err.Error()))
		os.Exit(0)
	}
	fmt.Print(out)

	if *verboseParam {
		p.PrintHeaderResponse()
	}
}

func PrettyJSON(str string) (string, error) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(str), "", "    "); err != nil {
		return "", err
	}

	resp := "```json\n" + pretty.String() + "\n```\n"

	return resp, nil
}
