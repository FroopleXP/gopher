package main

import (
    "log"
    "net"
    "fmt"
    "io"
    "os"
    "strings"
    "strconv"
    "bufio"
)

// Spec: https://www.rfc-editor.org/rfc/rfc1436

// ET - Element Type
type ET string

const (
    ETFile       ET = "0"
    ETDirectory  ET = "1"
    ETError      ET = "3"
    ETInfo       ET = "i"
    ETBinary     ET = "9"
    ETUnknown    ET = "?" // NOTE: This is not spec-compliant - just our way of denoting a weird type
)

type Element struct {
    Type     ET
    Value    string
    Selector string
    Host     string
    Port     int

    // External denotes, if an element is a link, if the link is on the same server or another
    // is is always 'false' if the element is not a link
    External bool 
}

func printUsage() {
    fmt.Print("Usage:\n\tgopher [host] [selector]\n")
}

func parseET(val rune) ET {
    switch val {
    case '0':
        return ETFile
    case '1':
        return ETDirectory
    case '9':
        return ETBinary
    case 'i':
        return ETInfo
    default:
        return ETUnknown
    }
}

func parseLine(line string) (*Element, error) {
    parts := strings.Split(line, "\t")
    if len(parts) < 4 {
        return nil, fmt.Errorf("invalid parts in line, got '%d' wanted 4", len(parts))
    }

    var e Element    
    e.Type     = parseET(rune(parts[0][0]))
    e.Value    = parts[0][1:]
    e.Selector = parts[1]
    e.Host     = parts[2]

    i, err := strconv.Atoi(parts[3])
    if err != nil {
        return nil, err
    }
    e.Port = i

    return &e, nil
}

func getPage(host string, selector string) ([]Element, error) {
    selector += "\r\n"

    els := make([]Element, 0)

    // If there's no ':' present, we assume the user hasn't supplied a port
    parts := strings.Split(host, ":")
    if len(parts) < 2 {
        host += ":70"
    }

    c, err := net.Dial("tcp", host)
    if err != nil {
        return els, err
    }
    defer c.Close()

    _, err = c.Write([]byte(selector))
    if err != nil {
        return els, err
    }
    
    r := bufio.NewReader(c)
    
    var e error = nil
    for {
        // TODO: 'isPrefix' is not handled here, see docs as to why this is important
        line, _, err := r.ReadLine()
        if err != nil {
            if err == io.EOF {
                break
            }
            e = err
            break
        }
    
        if len(line) == 0 {
            continue
        }

        el, err := parseLine(string(line))
        if err != nil {
            e = err
            break
        }

        el.External = el.Host == host

        els = append(els, *el)
    } 

    if e != nil {
        return els, e
    }

    return els, nil
}

func getFile(w io.Writer, host string, selector string) error {
    log.Println("getFile")

    log.Println("connecting to server")
    c, err := net.Dial("tcp", host)
    if err != nil {
        return err
    }
    
    log.Printf("writing selector '%s' to server\n", selector)
    _, err = c.Write([]byte(selector + "\r\n"))
    if err != nil {
        return err
    }

    log.Println("copying response from server to provided writer")
    n, err := io.Copy(w, c)
    if err != nil {
        return err
    }

    log.Printf("written %d byte(s) to writer\n", n)

    return nil
}

func printPage(elements []Element) {
    for _, el := range elements {
        switch el.Type {
        case ETInfo:
            fmt.Println(el.Value)
        case ETFile:
            fmt.Printf("%s\n", el.Value)
        case ETDirectory:
            fmt.Printf("%s...\n", el.Value)
        default:
            fmt.Printf("%s [?]\n", el.Value)
        } 
    }
}

func client(args []string) {
    if len(args) < 2 {
        fmt.Println("invalid client usage")
        return
    }

    url := args[0]
    sel := args[1]

    els, err := getPage(url, sel)
    if err != nil {
        log.Fatalf("failed to get page: %v\n", err)
    }

    printPage(els)
}

func main() {
    if len(os.Args) < 2 {
        log.Fatalf("expected subcommand [client|server]\n")
    }

    switch os.Args[1] {
    case "client":
        client(os.Args[2:])
    case "server":
        server(os.Args[2:])
    case "proxy":
        proxy(os.Args[2:])
    default:
        log.Fatalf("unknown subcommand '%s'\n", os.Args[1])    
    }
}
