package main

import (
	"encoding/json"
	"fmt"
	"github.com/rancher/hello-world/templates"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type NameRequest struct {
	Name string
}

const defaultListenPort = "80"

type HelloWorldConfig struct {
	Hostname string
	Services map[string]string
	Headers  http.Header
	Host     string
}

func (config *HelloWorldConfig) GetManifest() (string, error) {
	return templates.CompileTemplateFromMap(templates.HelloWorldTemplate, config)
}

func (config *HelloWorldConfig) getServices() {
	k8sServices := make(map[string]string)

	for _, evar := range os.Environ() {
		show := strings.Split(evar, "=")
		regName := regexp.MustCompile("^.*_PORT$")
		regLink := regexp.MustCompile("^(tcp|udp)://.*")
		if regName.MatchString(show[0]) && regLink.MatchString(show[1]) {
			k8sServices[strings.TrimSuffix(show[0], "_PORT")] = show[1]
		}
	}

	config.Services = k8sServices
}

func (config *HelloWorldConfig) Init(r *http.Request) {
	config.Hostname, _ = os.Hostname()
	config.Host = r.Host
	config.Headers = r.Header
	config.getServices()
}

func handler(w http.ResponseWriter, r *http.Request) {
	config := &HelloWorldConfig{}
	config.Init(r)
	data, err := config.GetManifest()
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprint(w, data)
}

func stdout(w http.ResponseWriter, r *http.Request) {
	var n NameRequest

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Name that was submitted: ", n.Name)
}

func main() {
	webPort := os.Getenv("HTTP_PORT")
	if webPort == "" {
		webPort = defaultListenPort
	}

	fmt.Println("Running http service at", webPort, "port")
	http.HandleFunc("/", handler)
	http.HandleFunc("/stdout", stdout)
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(os.Getenv("PWD")))))
	http.ListenAndServe(":"+webPort, nil)
}
