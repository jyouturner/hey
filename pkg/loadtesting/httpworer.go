package loadtesting

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HttpWorker struct {
	Url string
}

func NewHttpWorker(ctx context.Context, url string) HttpWorker {
	s := HttpWorker{
		Url: url,
	}
	return s
}

func (s HttpWorker) Process(ctx context.Context, input []string) (*string, error) {
	if len(input) != 2 {
		return nil, fmt.Errorf("expect two strings a row %v", input)
	}
	return postToWebhookIngest(s.Url, input[0], input[1])
}

func postToWebhookIngest(url string, body string, sig string) (*string, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("client: error creating http request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client: error making http request: %v", err)
	}
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("client: could not read response body: %v", err)
	}
	fmt.Printf("client: response body: %s\n", resBody)
	data := string(resBody)
	return &data, nil
}
