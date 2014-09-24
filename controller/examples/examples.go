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

type generator struct {
	conf        *config
	client      *http.Client
	resourceIds map[string]string
}

type example struct {
	name string
	req  *request
}

func main() {
	conf, err := loadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	client.Transport = &roundTripRecorder{roundTripper: client.Transport}

	e := &generator{
		conf:        conf,
		client:      client,
		resourceIds: make(map[string]string),
	}

	examples := []example{
		{"key_create", e.createKey()},
		{"key_get", e.getKey()},
		{"key_list", e.listKeys()},
		{"key_delete", e.deleteKey()},
		{"app_create", e.createApp()},
		{"app_get", e.getApp()},
		{"app_update", e.updateApp()},
		{"artifact_create", e.createArtifact()},
		{"release_create", e.createRelease()},
		{"artifact_list", e.listArtifacts()},
		{"release_list", e.listReleases()},
		{"app_delete", e.deleteApp()},
	}

	// TODO: hit all controller endpoints

	res := make(map[string]string)
	for i := range examples {
		example := examples[i]
		res[example.name] = requestMarkdown(example.req)
	}

	data, _ := json.Marshal(res)
	if len(os.Args) > 1 {
		ioutil.WriteFile(os.Args[1], data, 0644)
	} else {
		io.Copy(os.Stdout, bytes.NewReader(data))
	}
}

func generatePublicKey() (string, error) {
	key := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDPI19fkFmPNg3MGqJorFTbetPJjxlhLDUJFALYe5DyqW0lAnb2R7XvXzj+kRX9LkwOeQjf6nM4bcXbd/H3YPlMDc9JfDuSGlwvo0X8KUQ6PopgyfQ15GA+8YDgwYcBJowIXqAc52GVNnBUeoZzBKvNnsVjAw6KkTPS0aZ6KBZadtYx+Y1fJJBoygh/gtPZ/MQry3XQRvbKPa0iU34Wcx8pXx5QVFLHvyORczQlEVyq5qa5DT86CRR/wC4yH32hkNGalGXY7sZg0j4EY4AeD2yCcmsp7hTt4Ql4gRp3r04ye4DZ7epdXW2tp2vJ3IVn+l6BSNooBIfoD7ZdkUVce51z some-comment`
	return key, nil
}

func (e *generator) DoNewRequest(method, path string, header http.Header, body io.Reader) (*http.Response, error) {
	url := "http://" + e.conf.controllerDomain + path
	req, err := e.NewRequest(method, url, header, body)
	if err != nil {
		return nil, err
	}
	res, err := e.client.Do(req)
	return res, err
}

func (e *generator) NewRequest(method, url string, header http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if header != nil {
		req.Header = header
	}
	req.SetBasicAuth("", e.conf.controllerKey)
	return req, err
}

func (e *generator) createResource(path string, body io.Reader) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	return e.DoNewRequest("POST", path, header, body)
}

func (e *generator) createKey() *request {
	key, err := generatePublicKey()
	res, err := e.createResource("/keys", strings.NewReader(fmt.Sprintf(`{
    "key": "ssh-rsa %s"
  }`, key)))
	var k ct.Key
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&k); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	e.resourceIds["key"] = k.ID
	return getRequests()[0]
}

func (e *generator) getKey() *request {
	keyId := e.resourceIds["key"]
	res, err := e.DoNewRequest("GET", "/keys/"+keyId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) listKeys() *request {
	res, err := e.DoNewRequest("GET", "/keys", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) deleteKey() *request {
	keyId := e.resourceIds["key"]
	res, err := e.DoNewRequest("DELETE", "/keys/"+keyId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) createApp() *request {
	res, err := e.createResource("/apps", strings.NewReader(`{
    "name": "my-app"
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) getApp() *request {
	res, err := e.DoNewRequest("GET", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) updateApp() *request {
	res, err := e.createResource("/apps/my-app", strings.NewReader(`{
    "name": "my-app",
    "meta": {
      "bread": "with hemp"
    }
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) deleteApp() *request {
	res, err := e.DoNewRequest("DELETE", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) createArtifact() *request {
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
	e.resourceIds["artifact"] = a.ID
	return getRequests()[0]
}

func (e *generator) listArtifacts() *request {
	res, err := e.DoNewRequest("GET", "/artifacts", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}

func (e *generator) createRelease() *request {
	artifactId := e.resourceIds["artifact"]
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
	e.resourceIds["release"] = r.ID
	return getRequests()[0]
}

func (e *generator) listReleases() *request {
	res, err := e.DoNewRequest("GET", "/releases", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
	return getRequests()[0]
}
