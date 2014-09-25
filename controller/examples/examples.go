package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
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
		{"provider_create", e.createProvider},
		{"provider_get", e.getProvider},
		{"provider_list", e.listProviders},
		{"provider_resource_create", e.createProviderResource},
		{"provider_resource_get", e.getProviderResource},
		{"provider_resource_update", e.updateProviderResource},
		{"provider_resource_list", e.listProviderResources},
		{"app_delete", e.deleteApp},
	}

	// TODO: hit all controller endpoints

	res := make(map[string]string)
	for _, ex := range examples {
		ex.f()
		res[ex.name] = requestMarkdown(getRequests()[0])
	}

	var out io.Writer
	if len(os.Args) > 1 {
		out, err = os.OpenFile(os.Args[1], os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		out = os.Stdout
	}
	encoder := json.NewEncoder(out)
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

func (e *generator) createProvider() {
	res, err := e.createResource("/providers", strings.NewReader(fmt.Sprintf(`{
    "url": "discoverd+http://example",
    "name": "example provider %d"
  }`, rand.Intn(1000000))))
	if err != nil {
		log.Fatal(err)
	}
	var p ct.Provider
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&p); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	e.resourceIds["provider"] = p.ID
}

func (e *generator) getProvider() {
	providerId := e.resourceIds["provider"]
	res, err := e.DoNewRequest("GET", "/providers/"+providerId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) listProviders() {
	res, err := e.DoNewRequest("GET", "/providers", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) createProviderResource() {
	providerId := e.resourceIds["provider"]
	res, err := e.createResource("/providers/"+providerId+"/resources", strings.NewReader(`{
    "external_id": "some-id",
    "env": {
      "SOME": "ENV Vars"
    }
  }`))
	if err != nil {
		log.Fatal(err)
	}
	var r ct.Resource
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&r); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	e.resourceIds["provider_resource"] = r.ID
}

func (e *generator) getProviderResource() {
	providerId := e.resourceIds["provider"]
	resourceId := e.resourceIds["provider_resource"]
	res, err := e.DoNewRequest("GET", "/providers/"+providerId+"/resources/"+resourceId, nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) updateProviderResource() {
	providerId := e.resourceIds["provider"]
	resourceId := e.resourceIds["provider_resource"]
	res, err := e.createResource("/providers/"+providerId+"/resources/"+resourceId, strings.NewReader(`{
    "external_id": "some-id",
    "env": {
      "SOME": "ENV Vars",
      "More": "Stuff"
    }
  }`))
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}

func (e *generator) listProviderResources() {
	providerId := "be6ccfebf10b4e12a0a9aca196c650aa"
	res, err := e.DoNewRequest("GET", "/providers/"+providerId+"/resources", nil, nil)
	if err == nil {
		io.Copy(ioutil.Discard, res.Body)
	}
}
