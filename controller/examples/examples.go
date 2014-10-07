package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	cc "github.com/flynn/flynn/controller/client"
	ct "github.com/flynn/flynn/controller/types"
)

type generator struct {
	conf        *config
	client      *cc.Client
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

	client, err = cc.NewClient("http://"+conf.controllerDomain, conf.controllerKey)
	if err != nil {
		log.Fatal(err)
	}
	client.HTTP.Transport = &roundTripRecorder{roundTripper: &http.Transport{}}

	e := &generator{
		conf:        conf,
		client:      client,
		resourceIds: make(map[string]string),
	}

	go e.listenAndServe()

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

	// TODO: Use client lib
	// TODO: POST /apps/:app_id/routes
	// TODO: GET /apps/:app_id/routes
	// TODO: GET /apps/:app_id/routes/:routes_type/:routes_id
	// TODO: DELETE /apps/:app_id/routes/:routes_type/:routes_id
	// TODO: GET /apps/:app_id/resources
	// TODO: PUT /apps/:app_id/release
	// TODO: GET /apps/:app_id/release
	// TODO: PUT /apps/:app_id/formatinos/:release_id
	// TODO: GET /apps/:app_id/formatinos/:release_id
	// TODO: DELETE /apps/:app_id/formatinos/:release_id
	// TODO: GET /apps/:app_id/formatinos
	// TODO: POST /apps/:app_id/jobs
	// TODO: PUT /apps/:app_id/jobs/:job_id
	// TODO: DELETE /apps/:app_id/jobs/:job_id
	// TODO: GET /apps/:app_id/jobs
	// TODO: GET /apps/:app_id/jobs/:job_id/log (event-stream)

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

func (e *generator) listenAndServe() {
	http.HandleFunc("/providers/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("GET /providers\n")
		w.WriteHeader(200)
	})

	http.ListenAndServe(":"+e.conf.ourPort, nil)
}

func generatePublicKey() (string, error) {
	key := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDPI19fkFmPNg3MGqJorFTbetPJjxlhLDUJFALYe5DyqW0lAnb2R7XvXzj+kRX9LkwOeQjf6nM4bcXbd/H3YPlMDc9JfDuSGlwvo0X8KUQ6PopgyfQ15GA+8YDgwYcBJowIXqAc52GVNnBUeoZzBKvNnsVjAw6KkTPS0aZ6KBZadtYx+Y1fJJBoygh/gtPZ/MQry3XQRvbKPa0iU34Wcx8pXx5QVFLHvyORczQlEVyq5qa5DT86CRR/wC4yH32hkNGalGXY7sZg0j4EY4AeD2yCcmsp7hTt4Ql4gRp3r04ye4DZ7epdXW2tp2vJ3IVn+l6BSNooBIfoD7ZdkUVce51z some-comment`
	return key, nil
}

func (e *generator) createKey() {
	pubKey, err := generatePublicKey()
	key, err := e.client.CreateKey(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["key"] = key.ID
}

func (e *generator) getKey() {
	e.client.GetKey(e.resourceIds["key"])
}

func (e *generator) listKeys() {
	e.client.KeyList()
}

func (e *generator) deleteKey() {
	e.client.DeleteKey(e.resourceIds["key"])
}

func (e *generator) createApp() {
	app := &ct.App{Name: "my-app"}
	err := e.client.CreateApp(app)
	if err == nil {
		e.resourceIds["app"] = app.ID
	}
}

func (e *generator) getApp() {
	e.client.GetApp(e.resourceIds["app"])
}

func (e *generator) listApps() {
	e.client.AppList()
}

func (e *generator) updateApp() {
	app := &ct.App{
		Name: "my-app",
		Meta: map[string]string{
			"bread": "with hemp",
		},
	}
	e.client.CreateApp(app)
}

func (e *generator) deleteApp() {
	e.client.DeleteApp(e.resourceIds["app"])
}

func (e *generator) createArtifact() {
	artifact := &ct.Artifact{
		Type: "docker",
		URI:  "example://uri",
	}
	err := e.client.CreateArtifact(artifact)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["artifact"] = artifact.ID
}

func (e *generator) listArtifacts() {
	e.client.ArtifactList()
}

func (e *generator) createRelease() {
	release := &ct.Release{
		ArtifactID: e.resourceIds["artifact"],
		Env: map[string]string{
			"some": "info",
		},
		Processes: map[string]ct.ProcessType{
			"foo": ct.ProcessType{
				Cmd: []string{"ls", "-l"},
				Env: map[string]string{
					"BAR": "baz",
				},
			},
		},
	}
	err := e.client.CreateRelease(release)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["release"] = release.ID
}

func (e *generator) listReleases() {
	e.client.ReleaseList()
}

func (e *generator) createProvider() {
	t := time.Now().UnixNano()
	provider := &ct.Provider{
		Name: fmt.Sprintf("example provider %d", t),
		URL:  fmt.Sprintf("discoverd+http://%s/providers/%d", net.JoinHostPort(e.conf.ourAddr, e.conf.ourPort), t),
	}
	err := e.client.CreateProvider(provider)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["provider"] = provider.ID
}

func (e *generator) getProvider() {
	e.client.GetProvider(e.resourceIds["provider"])
}

func (e *generator) listProviders() {
	e.client.ProviderList()
}

func (e *generator) createProviderResource() {
	resourceReq := &ct.ResourceReq{
		ProviderID: e.resourceIds["provider"],
	}
	resource, err := e.client.ProvisionResource(resourceReq)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["provider_resource"] = resource.ID
}

func (e *generator) getProviderResource() {
	providerID := e.resourceIds["provider"]
	resourceID := e.resourceIds["provider_resource"]
	e.client.GetResource(providerID, resourceID)
}

func (e *generator) updateProviderResource() {
	resource := &ct.Resource{
		ID:         e.resourceIds["provider_resource"],
		ProviderID: e.resourceIds["provider"],
	}
	e.client.PutResource(resource)
}

func (e *generator) listProviderResources() {
	e.client.ResourceList(e.resourceIds["provider"])
}
