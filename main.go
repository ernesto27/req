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
	userAgent  string
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
	params := Params{}

	flag.StringVar(&params.typeP, "t", "http", "type connection (ws, gq, http, grpc)")
	flag.StringVar(&params.url, "u", "", "url to connect")
	flag.StringVar(&params.message, "p", "", "data send to server")
	flag.StringVar(&params.query, "q", "", "query params")
	flag.BoolVar(&params.verbose, "v", false, "show response server headers")
	flag.StringVar(&params.header, "h", "", "header params")
	flag.StringVar(&params.method, "m", "GET", "method request")
	flag.StringVar(&params.file, "f", "", "file path")
	flag.StringVar(&params.userAgent, "a", "", "user agent header")

	// GRPC exclusive flags
	flag.StringVar(&params.importPath, "import-path", "", "The path to a directory from which proto sources can be imported, for use with -proto flags")
	flag.StringVar(&params.proto, "proto", "", "The proto file")
	flag.StringVar(&params.methodName, "method", "", "The service method to call")

	download := flag.Bool("d", false, "download content")
	flag.Parse()

	if len(params.message) > 0 {
		atSymbol := 64
		if params.message[0] == byte(atSymbol) {
			filePath := string(params.message[1:])

			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println(styleErr.Render("Error reading file: " + err.Error()))
				return
			}
			params.message = string(content)
		}
	}

	validProtocos := []string{"ws", "gq", "http", "grpc"}

	valid := false
	for _, v := range validProtocos {
		if params.typeP == v {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Println(styleErr.Render("Invalid protocol, valid are: http, ws, gq"))
		return
	}

	p := getProtocol(params)
	defer p.Close()

	if params.message == "" {
		p.OnMessageReceived()
	}

	resp, err := p.RequestResponse()
	if err != nil {
		fmt.Println(styleErr.Render(err.Error()))
		return
	}

	if *download {
		err := p.Download()
		if err != nil {
			fmt.Println(styleErr.Render(err.Error()))
		}
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
		if params.verbose {
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

	if params.verbose {
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
