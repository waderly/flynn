package main

import (
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
	f    func()
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
		{"key_create", e.createKey},
		{"key_get", e.getKey},
		{"key_list", e.listKeys},
		{"key_delete", e.deleteKey},
		{"app_create", e.createApp},
		{"app_get", e.getApp},
		{"app_list", e.listApps},
		{"app_update", e.updateApp},
		{"artifact_create", e.createArtifact},
		{"release_create", e.createRelease},
		{"artifact_list", e.listArtifacts},
		{"release_list", e.listReleases},
		{"app_delete", e.deleteApp},
	}

	// TODO: hit all controller endpoints

	res := make(map[string]string)
	for _, ex := range examples {
		ex.f()
		res[ex.name] = requestMarkdown(getRequests()[0])
	}

	var encoder *json.Encoder
	if len(os.Args) > 1 {
		file, err := os.OpenFile(os.Args[1], os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		encoder = json.NewEncoder(file)
	} else {
		encoder = json.NewEncoder(os.Stdout)
	}
	encoder.Encode(res)
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

func (e *generator) createKey() {
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
}

func (e *generator) getKey() {
	keyId := e.resourceIds["key"]
	res, err := e.DoNewRequest("GET", "/keys/"+keyId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) listKeys() {
	res, err := e.DoNewRequest("GET", "/keys", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) deleteKey() {
	keyId := e.resourceIds["key"]
	res, err := e.DoNewRequest("DELETE", "/keys/"+keyId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) createApp() {
	res, err := e.createResource("/apps", strings.NewReader(`{
    "name": "my-app"
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) getApp() {
	res, err := e.DoNewRequest("GET", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) listApps() {
	res, err := e.DoNewRequest("GET", "/apps", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) updateApp() {
	res, err := e.createResource("/apps/my-app", strings.NewReader(`{
    "name": "my-app",
    "meta": {
      "bread": "with hemp"
    }
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) deleteApp() {
	res, err := e.DoNewRequest("DELETE", "/apps/my-app", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) createArtifact() {
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
}

func (e *generator) listArtifacts() {
	res, err := e.DoNewRequest("GET", "/artifacts", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) createRelease() {
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
}

func (e *generator) listReleases() {
	res, err := e.DoNewRequest("GET", "/releases", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}
