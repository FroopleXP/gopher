package main

import (
    "log"
    "net"
    "flag"
    "io"
    "fmt"
    "bufio"
    "path"
    "os"
)


func writeErrorResponse(w io.Writer, msg string) error {
    msg = fmt.Sprintf("3%s\t/\t(error)\t0\r\n", msg)
    if _, err := w.Write([]byte(msg)); err != nil {
        return err
    }
    return nil
}

func handleClient(c net.Conn) error {
    log.Printf("got new client connection (%s)\n", c.RemoteAddr().String())

    r := bufio.NewReader(c)

    var sel []byte = []byte{}

    var e error = nil
    sel, _, e = r.ReadLine()
    if e != nil && e != io.EOF {
        return e
    }

    // TODO: Build path relative to the serve directory, being careful not to 
    // allow a malicious cunt to request a file outside of that directory
    p := path.Join("./example", string(sel))

    s, err := os.Stat(p)
    if err != nil {
        return writeErrorResponse(c, fmt.Sprintf("path '%s' does not exist", p))
    }

    if s.IsDir() {
        p = path.Join(p, "gophermap")
    }
     
    file, err := os.Open(p)    
    if err != nil {
        return writeErrorResponse(c, fmt.Sprintf("path '%s' could not be opened, please contact the server administrator", p))
    }
    defer file.Close()

    r1 := bufio.NewReader(file)
    w1 := bufio.NewWriter(c)

    var buf []byte = make([]byte, 1024)
    e = nil 
        
    for {
        log.Printf("reading file\n")
        n, err := r1.Read(buf)
        if n > 0 {
            log.Printf("read %d byte(s) from file\n", n)
            if _, err := w1.Write(buf[:n]); err != nil {
                e = err
                break
            }
        }
        if err == io.EOF {
            log.Printf("reached 'eof' on reading selector '%s'\n", p)
            break
        }
        e = err
    }

    if e != nil {
        return e
    }

    log.Printf("writing final flush\n")

    return w1.Flush()
}

func server(args []string) {
    fs := flag.NewFlagSet("server", flag.ExitOnError)
    fServerListener := fs.String("listener", ":70", "listener address")

    if err := fs.Parse(args); err != nil {
        log.Printf("failed to parse server arguments: %v\n", err)
        return
    }

    log.Printf("starting server on listener: %s\n", *fServerListener)

    listener, err := net.Listen("tcp4", *fServerListener)
    if err != nil {
        log.Fatalf("failed to start tcp4 listener: %v\n", err)
    }

    for {
        c, err := listener.Accept()
        if err != nil {
            log.Printf("failed to accept client connection: %v\n", err)
            continue
        }
        
        // TODO: The whole client life-span should be maintained in 'handleClient'
        go func(c net.Conn) {
            if err := handleClient(c); err != nil {
                log.Printf("failed to handle client: %v\n", err)
                return
            }

            if err := c.Close(); err != nil {
                log.Printf("failed to close client connection: %v\n", err)
            }
        }(c)
    }
}



