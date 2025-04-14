package main

import (
    "log"
    "net"
    "fmt"
    "io"
    "os"
    "strings"
    "strconv"
    "net/url"
    "bufio"
)

const GopherServer string = "gopher.quux.org:70"

// ET - Element Type
type ET string

const (
    ETTextFile       ET = "0"
    ETGopherSubMenu  ET = "1"
    ETCCSONameServer ET = "2"
    ETError          ET = "3"
    ETBinHex         ET = "4"
    ETDos            ET = "5"
    ETUnencFile      ET = "6"
    ETGopherFTS      ET = "7"
    ETInfo           ET = "i"
    ETUnknown        ET = "?" // NOTE: This is not spec-compliant - just our way of denoting a weird type
)

type Element struct {
    Type     ET
    Value    string
    Selector string
    Host     string
    Port     int
}

func parseET(val rune) ET {
    switch val {
    case '0':
        return ETTextFile
    case '1':
        return ETGopherSubMenu
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

func getPage(page string, selector string) ([]Element, error) {
    selector += "\r\n"

    els := make([]Element, 0)

    u, err := url.Parse(page)
    if err != nil {
        return els, nil
    }

    if u.Scheme != "gopher" {
        return els, fmt.Errorf("invalid url scheme '%s'", u.Scheme)
    }
    
    // TODO: Only default to 70, respect what the user may have put
    page = u.Host + ":70"

    c, err := net.Dial("tcp", page)
    if err != nil {
        return els, err
    }
    defer c.Close()

    _, err = c.Write([]byte(selector))
    if err != nil {
        return els, err
    }

    // TODO: Do a flush here? For good measure?
    
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

        el, err := parseLine(string(line))
        if err != nil {
            e = err
            break
        }

        els = append(els, *el)
    } 

    if err != nil {
        return els, e
    }

    return els, nil
}

func printPage(elements []Element) {
    for _, el := range elements {
        switch el.Type {
        case ETInfo:
            fmt.Println(el.Value)
            break
        case ETUnknown:
            log.Printf("got unknown element type: '%s'\n", el.Type)
            break
        default:
            fmt.Printf("%s [%s]\n", el.Value, el.Selector) 
        } 
    }
}

func main() {
    if len(os.Args) < 3 {
        log.Fatalf("invalid arguments")
    }

    url := os.Args[1]
    sel := os.Args[2]

    els, err := getPage(url, sel)
    if err != nil {
        log.Fatalf("failed to get page: %v\n", err)
    }

    printPage(els)
}
