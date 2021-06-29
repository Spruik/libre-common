package commandLine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type DaemonCLI struct {
	port string

	BaseURL   *url.URL
	UserAgent string

	httpClient *http.Client
}

func NewDaemonCLI(port string, userAgentName string) *DaemonCLI {
	baseUrl, err := url.Parse(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		panic(err)
	}
	return &DaemonCLI{
		port:      port,
		BaseURL:   baseUrl,
		UserAgent: userAgentName,
		httpClient: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		},
	}
}

func (s *DaemonCLI) Init() {
	req, err := http.NewRequest("GET", s.BaseURL.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.UserAgent)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var eps = struct {
		DaemonName string
		Endpoints  []string
	}{}
	err = json.NewDecoder(resp.Body).Decode(&eps)

}

func (s *DaemonCLI) Converse() {
	var quit bool = false
	reader := bufio.NewReader(os.Stdin)
	for !quit {
		fmt.Printf("%s >", s.UserAgent)
		// ReadString will block until the delimiter is entered
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occured while reading input. Please try again", err)
			continue
		}

		// remove the delimeter from the string
		input = strings.TrimSuffix(input, "\n")
		fmt.Println(input)

		if input == "quit" {
			quit = true
		}

		s.doCommandFromString(input)
	}
}

func (s *DaemonCLI) doCommandFromString(input string) {
	tokens := strings.Split(input, " ")
	if len(tokens) > 0 {
		switch tokens[0] {
		case "loggers":
		case "":
		}
	}
}
