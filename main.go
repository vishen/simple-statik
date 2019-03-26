package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	dir        = flag.String("dir", "templates/", "directory to serve files from")
	configFile = flag.String("config", "", "config file")
	port       = flag.String("port", "8081", "port to use for http server")
)

type kv struct {
	key   string
	value string
}

type route struct {
	urls   []string
	prefix bool

	httpStatusCode int
	httpHeaders    []kv

	file    string
	folder  string
	message string
}

func removeComment(line string) string {
	for i, c := range line {
		if c == '#' {
			return line[0:i]
		}
	}
	return line
}

func parseRoutes(config string) ([]route, error) {
	routes := []route{}
	r := route{}
	for i, line := range strings.Split(config, "\n") {
		// Ignore blank lines or comments.
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		line = removeComment(line)
		if strings.HasPrefix(line, "->") {
			line = strings.Replace(line, "->", "", 1)
			keyAndValue := strings.SplitN(strings.TrimSpace(line), "=", 2)
			switch val := strings.Trim(keyAndValue[1], "\""); keyAndValue[0] {
			case "file":
				r.file = val
			case "folder":
				r.folder = val
			case "prefix":
				r.prefix = val == "true"
			case "message":
				r.message = val
			case "http_status_code":
				var err error
				r.httpStatusCode, err = strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf(":%d invalid http status code %q, needs to be a number: %v", i+1, val, err)
				}
			case "header":
				keyAndValue := strings.Split(val, "=")
				if len(keyAndValue) != 2 {
					return nil, fmt.Errorf(":%d invalid http header %q", i+1, val)
				}
				r.httpHeaders = append(r.httpHeaders, kv{
					key:   keyAndValue[0],
					value: keyAndValue[1],
				})
			default:
				return nil, fmt.Errorf(":%d unknown configuration %q", i+1, line)
			}
		} else {
			if len(r.urls) > 0 {
				// TODO: make sure some config settings are set
				routes = append(routes, r)
				r = route{}
			}
			for _, url := range strings.Split(line, " ") {
				// TODO: Check that the url is a valid url string
				url = strings.TrimSpace(url)
				r.urls = append(r.urls, url)
			}
		}
	}
	// Add the last route.
	// TODO: make sure some config settings are set, add
	// a validate function?
	routes = append(routes, r)
	return routes, nil
}

type server struct {
	routes []route
	dir    string
}

func newServer(routes []route, dir string) *server {
	return &server{
		routes: routes,
		dir:    dir,
	}
}

func (s server) findRoute(path string) (route, bool) {
	for _, r := range s.routes {
		for _, url := range r.urls {
			if r.prefix && strings.HasPrefix(path, url) {
				return r, true
			} else if url == path {
				return r, true
			} else if url == "_" {
				return r, true
			}
		}
	}
	return route{}, false
}

func (s *server) routeHandler(w http.ResponseWriter, r *http.Request) {
	route, found := s.findRoute(r.URL.Path)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "simple-statik didn't find route")
		return
	}

	// Set any http header.
	for _, h := range route.httpHeaders {
		w.Header().Set(h.key, h.value)
	}

	writeHeader := func() {
		// Set the response status code if set.
		if route.httpStatusCode > 0 {
			w.WriteHeader(route.httpStatusCode)
		}
	}

	serveFile := func(filename string) bool {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("unable to open file %s: %v", filename, err)
			w.WriteHeader(404)
			return false
		}
		writeHeader()
		buf := bytes.NewBuffer(b)
		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("unable to write file to response: %v", err)
			return false
		}
		return true
	}

	switch {
	case route.message != "":
		writeHeader()
		fmt.Fprintf(w, route.message)
	case route.file != "":
		if served := serveFile(filepath.Join(s.dir, route.file)); !served {
			return
		}
	case route.folder != "":
		path := filepath.Join(s.dir, route.folder)
		path = filepath.Join(path, strings.Replace(r.URL.Path, route.folder, "", 1))
		if served := serveFile(path); !served {
			return
		}
	default:
		writeHeader()
	}
}

func getPort() string {
	p := os.Getenv("PORT")
	if p == "" {
		p = *port
		if p == "" {
			log.Fatal("missing -port flag or 'PORT' env")
		}
	}
	return ":" + p
}

func getConfig() string {
	cf := *configFile
	if cf == "" {
		cf = os.Getenv("CONFIG_FILE")
		if cf == "" {
			log.Fatal("missing -config flag or 'CONFIG_FILE' env")
		}
	}
	config, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	return string(config)
}

func getDir() string {
	d := os.Getenv("DIR")
	if d == "" {
		d = *dir
		if d == "" {
			log.Fatal("missing -dir flag or 'DIR' env")
		}
	}
	return d
}

func main() {
	flag.Parse()

	routes, err := parseRoutes(getConfig())
	if err != nil {
		log.Fatal(err)
	}

	s := newServer(routes, getDir())

	p := getPort()
	http.HandleFunc("/", s.routeHandler)
	fmt.Printf("listening on %s\n", p)
	log.Fatal(http.ListenAndServe(p, nil))
}
