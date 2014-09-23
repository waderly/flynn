package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	ct "github.com/flynn/flynn/controller/types"
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
	e.getApp()
	e.updateApp()

	artifactId := e.createArtifact()
	e.createRelease(artifactId)

	e.listArtifacts()
	e.listReleases()

	e.deleteApp()

	// TODO: hit all controller endpoints

	res := make(map[string]string)
	for k, v := range e.examples {
		res[k] = requestMarkdown(v)
	}

	data, _ := json.Marshal(res)
	if len(os.Args) > 1 {
		ioutil.WriteFile(os.Args[1], data, 0644)
	} else {
		io.Copy(os.Stdout, bytes.NewReader(data))
	}
}

func (e *exampler) DoNewRequest(method, path string, header http.Header, body io.Reader) (*http.Response, error) {
	url := "http://" + e.conf.controllerDomain + path
	req, err := e.NewRequest(method, url, header, body)
	if err != nil {
		return nil, err
	}
	res, err := e.client.Do(req)
	return res, err
}

func (e *exampler) NewRequest(method, url string, header http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if header != nil {
		req.Header = header
	}
	req.SetBasicAuth("", e.conf.controllerKey)
	return req, err
}

func (e *exampler) createResource(path string, body io.Reader) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return e.DoNewRequest("POST", path, header, body)
}

func (e *exampler) createApp() {
	res, err := e.createResource("/apps", strings.NewReader(`{
    "name": "my-app"
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["app_create"] = getRequests()[0]
}

func (e *exampler) getApp() {
	res, err := e.DoNewRequest("GET", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["app_get"] = getRequests()[0]
}

func (e *exampler) updateApp() {
	res, err := e.createResource("/apps/my-app", strings.NewReader(`{
    "name": "my-app",
    "meta": {
      "bread": "with hemp"
    }
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["app_update"] = getRequests()[0]
}

func (e *exampler) deleteApp() {
	res, err := e.DoNewRequest("DELETE", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["app_delete"] = getRequests()[0]
}

func (e *exampler) createArtifact() string {
	res, err := e.createResource("/artifacts", strings.NewReader(`{
    "type": "docker",
    "uri": "example://uri"
  }`))
	if err != nil {
		log.Fatal(err)
	}
	var a ct.Artifact
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&a); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	e.examples["artifact_create"] = getRequests()[0]
	return a.ID
}

func (e *exampler) listArtifacts() {
	res, err := e.DoNewRequest("GET", "/artifacts", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["artifact_list"] = getRequests()[0]
}

func (e *exampler) createRelease(artifactId string) string {
	res, err := e.createResource("/releases", strings.NewReader(fmt.Sprintf(`{
    "artifact": "%s",
    "env": {
      "some": "info"
    },
    "processes": {
      "foo": {
        "cmd": ["ls", "-l"],
        "env": {
          "BAR": "baz"
        }
      }
    }
  }`, artifactId)))
	if err != nil {
		log.Fatal(err)
	}
	var r ct.Release
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&r); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	e.examples["release_create"] = getRequests()[0]
	return r.ID
}

func (e *exampler) listReleases() {
	res, err := e.DoNewRequest("GET", "/releases", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	e.examples["release_list"] = getRequests()[0]
}
