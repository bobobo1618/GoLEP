package main

import (
    "errors"
    "flag"
    "io"
    "log"
    "net"
    "net/http"
    "net/url"
    "os"
    "path"
    "regexp"
    "time"
)

var root = flag.String("root", ".", "Root path to serve files from.")
var listenAddr = flag.String("listen", ":8080", "Host and port to listen on.")
var readTimeout = flag.Int("readTimeout", 10, "Timeout for reading.")
var writeTimeout = flag.Int("writeTimeout", 10, "Timeout for writing.")
var leptonUrlString = flag.String("leptonSocket", "tcp://localhost:2402", "Socket to use to connect to Lepton. e.g. tcp://localhost:2402, unix:///tmp/.leptonsock")
var leptonUrl *url.URL
var jpegExp = regexp.MustCompile(`^.*\.jpe?g$`)
var lepExp = regexp.MustCompile(`^.*\.lep$`)

type leptonHandler struct {
    http.Handler
}

func (leptonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        filePath := path.Join(*root, r.URL.Path)
        file, err := os.Open(filePath)
        
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            io.WriteString(w, err.Error())
            log.Printf("%s not found.\n", filePath)
            return
        }
        defer file.Close()

        var tcpConn *net.TCPConn
        var unixConn *net.UnixConn

        // Assume that if there's no host, to connect as a UNIX socket
        if leptonUrl.Scheme == "tcp" {
            addr, addrErr := net.ResolveTCPAddr(leptonUrl.Scheme, leptonUrl.Host)
            if addrErr != nil {
                err = addrErr
            }
            tcpConn, err = net.DialTCP(leptonUrl.Scheme, nil, addr)
        } else if leptonUrl.Scheme == "unix" {
            addr, addrErr := net.ResolveUnixAddr(leptonUrl.Scheme, leptonUrl.Path)
            if addrErr != nil {
                err = addrErr
            }
            unixConn, err = net.DialUnix(leptonUrl.Scheme, nil, addr)
        } else {
            err = errors.New("No valid URL scheme specified.")
        }
        if err != nil {
            w.WriteHeader(http.StatusBadGateway)
            io.WriteString(w, "Couldn't connect to Lepton.")
            log.Printf("Couldn't connect to %s\n", leptonUrlString)
            return
        }
        firstReadBuffer := make([]byte, 100)
        var numRead int

        if leptonUrl.Scheme == "tcp" {
            io.Copy(tcpConn, file)
            tcpConn.CloseWrite()
            numRead, err = tcpConn.Read(firstReadBuffer)
        } else if leptonUrl.Scheme == "unix" {
            io.Copy(unixConn, file)
            unixConn.CloseWrite()
            numRead, err = unixConn.Read(firstReadBuffer)
        }

        if err != nil || numRead == 0 {
            w.WriteHeader(http.StatusInternalServerError)
            io.WriteString(w, "Lepton returned nothing or a TCP error occurred.")
            log.Printf("Bad Lepton response: %d, %s\n", numRead, err.Error())
            return
        }

        if jpegExp.MatchString(r.URL.Path) {
            w.Header().Add("Content-Type", "image/lepton")
        } else if lepExp.MatchString(r.URL.Path) {
            w.Header().Add("Content-Type", "image/jpeg")
        }

        w.WriteHeader(http.StatusOK)
        w.Write(firstReadBuffer)
        if leptonUrl.Scheme == "tcp" {
            io.Copy(w, tcpConn)
            tcpConn.Close()
        } else if leptonUrl.Scheme == "unix" {
            io.Copy(w, unixConn)
            unixConn.Close()
        }
        return
    }
}

func main() {
    flag.Parse()
    var err error
    leptonUrl, err = url.Parse(*leptonUrlString)
    if err != nil {
        log.Fatal("Error parsing Lepton URL: %s", err)
    }
    s := &http.Server{
        Addr: *listenAddr,
        Handler: leptonHandler{},
        ReadTimeout: time.Duration(*readTimeout) * time.Second,
        WriteTimeout: time.Duration(*writeTimeout) * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    log.Fatal(s.ListenAndServe())
}