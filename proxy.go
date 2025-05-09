package main

import (
    "log"
    "flag"
    "fmt"
    "os"
    "net/http"
    "html/template"
	"strings"
	"path"
)

func printProxyUsage() {
    fmt.Println("usage: proxy {...opts} <server>")
}

func proxy(args []string) {
    fs := flag.NewFlagSet("proxy", flag.ExitOnError)
    fProxyListener := fs.String("listener", ":8080", "proxy listener address")

    if err := fs.Parse(args); err != nil {
        log.Fatalf("failed to parse proxy arguments: %v\n", err)
    }

    fs.Usage = printProxyUsage
    
    if fs.NArg() < 1 {
        fmt.Println("invalid argument(s)")
        fs.Usage()
        os.Exit(1)
    }

    server := fs.Args()[0]

    templ, err := template.ParseFiles("./template/gophermap.html")
    if err != nil {
        log.Fatalf("failed to parse html templates: %v\n", err)
    }

    log.Printf("starting proxy server on listener :%s\n", *fProxyListener)
    log.Printf("proxy starts from %s\n", server)

    http.HandleFunc("GET /{url...}", func (w http.ResponseWriter, r *http.Request) {
		typ := r.URL.Query().Get("type")
		if typ == "" {
			typ = string(ETDirectory)
		}

		t := rune(typ[0])

        url := r.PathValue("url")
        if url == "" {
            w.WriteHeader(404)
            return
        }

		// BUG: If a user requests 'gopher.quux.org:70' without the trailing
		// '/', this won't work. 
		host := ""
		selector := "/"

		parts := strings.SplitN(url, "/", 2)
		if len(parts) == 2 {
			host     = parts[0]
			selector = parts[1]
		}

		log.Printf("proxying request to '%s', for selector '%s'\n", host, selector)
    
        // TODO: There's lots of duplication here
		if t == ETFile {
			w.Header().Set("Content-Type", "text/plain")
			if err := getFile(w, host, selector); err != nil {
				log.Printf("failed to get file: %v\n", err)
                w.WriteHeader(503)
                return
			}
			return

		} else if t == ETBinary {
			binName := path.Base(selector)
			if binName == "" {
				binName = "unknown.bin"
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename=\"" + binName + "\"")

			if err := getFile(w, host, selector); err != nil {
				log.Printf("failed to get file: %v\n", err)
                w.WriteHeader(503)
                return
			}
			return

		} else if t == ETGif {
			w.Header().Set("Content-Type", "image/gif")
			if err := getFile(w, host, selector); err != nil {
				log.Printf("failed to get file: %v\n", err)
                w.WriteHeader(503)
                return
			}
			return

		} else if t == ETDirectory {
			els, err := getPage(host, selector)
			if err != nil {
				log.Printf("failed to get directory '%s': %v\n", selector, err)
				w.WriteHeader(503)
				return
			}
			if err := templ.ExecuteTemplate(w, "gophermap.html", els); err != nil {
				log.Printf("failed to execute template: %v\n", err)
				w.WriteHeader(500)
				return
			}
			return
		}

		w.WriteHeader(401)
		log.Printf("got unknown request type '%c'\n", t)
		return

    })

    if err := http.ListenAndServe(*fProxyListener, nil); err != nil {
        log.Fatalf("failed to start http proxy: %v\n", err)
    }
}
