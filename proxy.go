package main

import (
    "log"
    "flag"
    "fmt"
    "os"
    "net/http"
    "html/template"
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

    http.HandleFunc("GET /", func (w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Location", fmt.Sprintf("/%s", server))
        w.WriteHeader(302)
    })

    http.HandleFunc("GET /{url}", func (w http.ResponseWriter, r *http.Request) {
        url := r.PathValue("url")
        if url == "" {
            w.WriteHeader(404)
            return
        }

        els, err := getPage(url, "/")
        if err != nil {
            log.Printf("failed to execute template: %v\n", err)
            w.WriteHeader(500)
            return
        }
    
        if err := templ.ExecuteTemplate(w, "gophermap.html", els); err != nil {
            log.Printf("failed to execute template: %v\n", err)
            w.WriteHeader(500)
            return
        }
    })

    if err := http.ListenAndServe(*fProxyListener, nil); err != nil {
        log.Fatalf("failed to start http proxy: %v\n", err)
    }
}
