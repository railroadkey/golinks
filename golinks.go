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
        shortname string
        url       string
        requests  int32
}

type settings struct {
        redirects []*redirect
        filename  string
}

func main() {
        port := flag.String("http_port", "8080", "Port number to listen")
        configFile := flag.String("config", "redirects.json", "Configuration filename")
        flag.Parse()
        log.Printf("Starting golinks...")
        r := newRedirector(*configFile)
        r.readConfig()

        http.HandleFunc("/add/", r.addLink)
        http.HandleFunc("/list/", r.getLinks)
        http.HandleFunc("/del/", r.delLink)
        http.HandleFunc("/", r.redirect)
        if err := http.ListenAndServe(":"+*port, nil); err != nil {
                log.Fatal("Unable to listen %s", err)
        }
}

func newRedirector(file string) *settings {
        return &settings{
                filename: file,
        }
}

func (m *settings) readConfig() {
        log.Printf("Reading configuration...")
        jsonBlob, err := ioutil.ReadFile(m.filename)
        if err != nil {
                log.Printf("No config file found. Using new config file.")
                return
        }
        if err := json.Unmarshal(jsonBlob, &m.redirects); err != nil {
                log.Printf("Error unmarshalling %s", err)
        }
}
func (m *settings) saveToDisk() error {
        b, err := json.Marshal(m.redirects)
        if err != nil {
                log.Printf("Error marshalling %s", err)
                return fmt.Errorf("Error marshalling %s", err)
        }

        if err := ioutil.WriteFile(m.filename, b, 0644); err != nil {
                // This function should probably return an error instead
                return fmt.Errorf("Unable to open file %s", err)
        }
        log.Printf("Saving to disk.")
        return nil
}

func (m *settings) getLinks(w http.ResponseWriter, r *http.Request) {
        var s string
        for _, v := range m.redirects {
                s += v.shortname + "->" + v.url + "<BR>"
        }
        sendHtml(w, s)
}

func (m *settings) redirect(w http.ResponseWriter, r *http.Request) {
        var url string
        req := strings.Split(r.URL.Path[1:], "/")
        args := strings.Join(req[1:], "/")
        sh := strings.Trim(req[0], " ")
        for _, v := range m.redirects {
                if v.shortname == sh {
                        url = v.url
                        h := w.Header()
                        h.Set("Cache-Control", "private, no-cache")
                        http.Redirect(w, r, url+"/"+args, 302)
                        break
                }
        }
        sendHtml(w, "Shortname "+sh+" not found!")

}

func (m *settings) delLink(w http.ResponseWriter, r *http.Request) {
        req := strings.Trim(r.URL.Path[5:], " ")
        for i, v := range m.redirects {
                if v.shortname == req {
                        m.redirects = append(m.redirects[:i], m.redirects[i+1:]...)
                        m.saveToDisk()
                        break
                }
        }
        http.Redirect(w, r, "/list", 302)
}

func (m *settings) addLink(w http.ResponseWriter, r *http.Request) {
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
                        if v.shortname == req[0] {
                                sendHtml(w, "Shortname already points to "+v.url)
                                return
                        }
                }
                // Add shortname, redirect to list.
                m.redirects = append(m.redirects, &redirect{
                        shortname: req[0],
                        url:       "http://" + req[1],
                        requests:  0,
                })
                err := m.saveToDisk()
                if err != nil {
                        http.Error(w, fmt.Sprintf("Internal error saving redirect."), http.StatusInternalServerError)
                }
                sendHtml(w, "Setting up redirect for  "+req[0]+" -> "+req[1])
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
                                                                                                                160,1         Bot
