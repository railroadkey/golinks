package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type redirect struct {
	Shortname string
	Url       string
	Requests  int32
}

type settings struct {
	redirects []*redirect
	filename  string
}

func main() {
	port := flag.String("http_port", "8080", "Port number to listen")
	configFile := flag.String("config", "redirects.json", "Port number to listen")
	flag.Parse()
	log.Printf("Starting golinks...")
	r := newRedirector(*configFile)
	r.ReadConfig()

	http.HandleFunc("/add/", r.AddLink)
	http.HandleFunc("/list/", r.GetLinks)
	http.HandleFunc("/del/", r.DelLink)
	http.HandleFunc("/", r.redirect)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("Unable to listen %s", err)
	}
	for {
	}
}

func newRedirector(file string) *settings {
	return &settings{
		filename: file,
	}
}

func (m *Settings) ReadConfig() {
	log.Printf("reading configuration...")
	jsonBlob, err := ioutil.ReadFile(m.Filename)
	if err != nil {
		log.Printf("Unable to read configuration %s", err)
		return
	}
	if err := json.Unmarshal(jsonBlob, &m.redirects); err != nil {
		log.Printf("error unmarshalling %s", err)
	}
}

func (m *settings) SaveToDisk() {
	b, err := json.Marshal(m.redirects)
	if err != nil {
		log.Printf("error marshalling %s", err)
	}

	if err := ioutil.WriteFile(m.Filename, b, 0644); err != nil {
		log.Fatalf("unable to open file %s", err)
	}
	log.Printf("saving to disk.")
}

func (m *settings) GetLinks(w http.ResponseWriter, r *http.Request) {
	var s string
	for _, v := range m.redirects {
		s += v.Shortname + "->" + v.Url + "<BR>"
	}
	sendHtml(w, s)
}

func (m *settings) redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	req := strings.Split(r.URL.Path[1:], "/")
	args := strings.Join(req[1:], "/")
	sh := strings.Trim(req[0], " ")
	for _, v := range m.redirects {
		if v.Shortname == sh {
			url = v.Url
			h := w.Header()
			h.Set("Cache-Control", "private, no-cache")
			http.redirect(w, r, url+"/"+args, 302)
			break
		}
	}
	sendHtml(w, "Shortname "+sh+" not found!")

}

func (m *settings) DelLink(w http.ResponseWriter, r *http.Request) {
	req := strings.Trim(r.URL.Path[5:], " ")
	for i, v := range m.redirects {
		if v.Shortname == req {
			m.redirects = append(m.redirects[:i], m.redirects[i+1:]...)
			m.SaveToDisk()
			break
		}
	}
	http.redirect(w, r, "/list", 302)
}

func (m *settings) AddLink(w http.ResponseWriter, r *http.Request) {
	req := strings.Split(r.URL.Path[5:], "|")

	// Sanitize input.
	for i, _ := range req {
		req[i] = strings.Trim(req[i], " ")
	}

	if len(req) >= 2 {
		// Ensure proper formatting for redirect url.
		var validUrl = regexp.MustCompile(`^[a-z, ,A-Z,0-9,-,_,/,:,\.]+$`)
		if !validUrl.MatchString(req[1]) {
			sendHtml(w, "Redirect url should be fully qualified URL http://something")
			return
		}
		// Verify if shortname already exists.
		for _, v := range m.redirects {
			if v.Shortname == req[0] {
				sendHtml(w, "Shortname already points to "+v.Url)
				return
			}
		}
		// Add shortname, redirect to list.
		m.redirects = append(m.redirects, &redirect{
			Shortname: req[0],
			Url:       "http://" + req[1],
			Requests:  0,
		})
		sendHtml(w, "Setting up redirect for  "+req[0]+" -> "+req[1])
		m.SaveToDisk()
		return
	}
	fmt.Fprintf(w, "Incorrect add format use add/<shortname>|url")
}

func sendHtml(w http.ResponseWriter, text string) {
	fmt.Fprintf(w, `<html>
				<head>
				<title>Redirects Setup</title>
				</head>
				<body>`)
	fmt.Fprintf(w, text)
	fmt.Fprintf(w, `</body>
				</html>`)
}
