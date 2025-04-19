package main

import (
    "log"
    "net"
    "flag"
    "fmt"
    "io"
    "bufio"
)

func handleClient(c net.Conn) error {
    log.Printf("got new client connection (%s)\n", c.RemoteAddr().String())

    r := bufio.NewReader(c)

    var sel []byte = []byte{}

    var e error = nil
    sel, _, e = r.ReadLine()
    if e != nil && e != io.EOF {
        return e
    }

    res := []byte(fmt.Sprintf("iWelcome, '%s'\t%s\t/\t70\n.\r\n", sel, sel))

    log.Print(string(res))
    
    if _, err := c.Write(res); err != nil {
        return err
    }

    return e
}

func server(args []string) {
    fs := flag.NewFlagSet("server", flag.ExitOnError)
    fServerListener := fs.String("listener", ":70", "listener address")

    if err := fs.Parse(args); err != nil {
        log.Printf("failed to parse server arguments: 5v\n", err)
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

        go func(c net.Conn) {
            if err := handleClient(c); err != nil {
                log.Printf("failed to handle client: %v\n", err)
                return
            }
        }(c)
    }
}



