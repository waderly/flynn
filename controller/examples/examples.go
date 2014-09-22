package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type exampler struct {
	conf     *config
	client   *http.Client
	examples map[string]*request
}

func main() {
	conf, err := loadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	client.Transport = &roundTripRecorder{roundTripper: client.Transport}

	e := &exampler{
		conf:     conf,
		client:   client,
		examples: make(map[string]*request),
	}

	e.createApp()

	// TODO: hit all controller endpoints

	res := make(map[string]string)
	for k, v := range e.examples {
		res[k] = requestMarkdown(v)
	}

	data, _ := json.Marshal(res)
	io.Copy(os.Stdout, bytes.NewReader(data)) // TODO: write to file
}

func (e *exampler) DoNewRequest(method, url string, header http.Header, body io.Reader) (*http.Response, error) {
	req, err := e.NewRequest(method, url, header, body)
	if err != nil {
		return nil, err
	}
	return e.client.Do(req)
}

func (e *exampler) NewRequest(method, url string, header http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if header != nil {
		req.Header = header
	}
	req.SetBasicAuth("", e.conf.controllerKey)
	return req, err
}

func (e *exampler) createApp() {
	url := fmt.Sprintf("http://%s/apps", e.conf.controllerDomain)
	body := strings.NewReader(`{ "name": "my-app" }`)
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	e.DoNewRequest("POST", url, header, body)
	e.examples["app_create"] = getRequests()[0]
}
