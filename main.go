package main

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itchyny/gojq"
	"github.com/pschou/gnuflag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var version = "debug"

var debug = false

var Headers = map[string]string{
	"content-type": "application/json",
}

type headerValue string

func (h *headerValue) Set(val string) error {
	parts := strings.SplitN(val, ":", 2)
	if len(parts) < 2 {
		return errors.New("Malformatted header")
	}
	Headers[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimPrefix(parts[1], " ")
	return nil
}
func (h *headerValue) Get() interface{} { return "" }
func (h *headerValue) String() string   { return "\"content-type: application/json\"" }

func main() {
	var maxTries int
	var delay, maxAge time.Duration
	var debug, raw, includeHeader, certIgnore, flush, useCache, followRedirects bool
	var cert, key, ca, cacheDir, method, postData string
	var headerVals *headerValue
	gnuflag.Default = "Default"
	gnuflag.Var(headerVals, "H", "Custom header to pass to server", "\"Key: Value\"")
	gnuflag.Var(headerVals, "header", "Custom header to pass to server", "\"Key: Value\"")
	gnuflag.BoolVar(&followRedirects, "L", false, "Follow redirects")
	gnuflag.BoolVar(&followRedirects, "location", false, "Follow redirects")
	gnuflag.BoolVar(&flush, "flush", false, "Force redownload, when using cache")
	gnuflag.BoolVar(&useCache, "C", false, "Use local cache to speed up static queries")
	gnuflag.BoolVar(&useCache, "cache", false, "Use local cache to speed up static queries")
	gnuflag.BoolVar(&debug, "debug", false, "Debug / verbose output")
	gnuflag.IntVar(&maxTries, "maxtries", 30, "Maximum number of tries", "TRIES")
	gnuflag.BoolVar(&raw, "r", false, "Raw output, no quotes for strings")
	gnuflag.BoolVar(&raw, "raw-output", false, "Raw output, no quotes for strings")
	gnuflag.BoolVar(&includeHeader, "i", false, "Include header in output")
	gnuflag.BoolVar(&includeHeader, "include", false, "Include header in output")
	gnuflag.BoolVar(&certIgnore, "k", false, "Ignore certificate validation checks")
	gnuflag.BoolVar(&certIgnore, "insecure", false, "Ignore certificate validation checks")
	gnuflag.StringVar(&method, "X", "GET", "Method to use for HTTP request (ie: POST/GET)", "METHOD")
	gnuflag.StringVar(&method, "request", "GET", "Method to use for HTTP request (ie: POST/GET)", "METHOD")
	gnuflag.StringVar(&postData, "d", "", "Data to use in POST (use @filename to read from file)", "STRING")
	gnuflag.StringVar(&postData, "data", "", "Data to use in POST (use @filename to read from file)", "STRING")
	gnuflag.StringVar(&ca, "cacert", "", "Use certificate authorities, PEM encoded", "FILE")
	gnuflag.StringVar(&cert, "cert", "", "Use client cert in request, PEM encoded", "FILE")
	gnuflag.StringVar(&cert, "E", "", "Use client cert in request, PEM encoded", "FILE")
	gnuflag.StringVar(&key, "key", "", "Key file for client cert, PEM encoded", "FILE")
	gnuflag.StringVar(&cacheDir, "cachedir", "/dev/shm", "Path for cache", "DIR")
	gnuflag.DurationVar(&delay, "retry-delay", 7*time.Second, "Delay between retries", "DURATION")
	gnuflag.DurationVar(&maxAge, "maxage", 4*time.Hour, "Max age for cache", "DURATION")

	gnuflag.Usage = func() {
		fmt.Println("Simple JSON URL downloader and parser tool, Written by paul (paulschou.com), Docs: github.com/pschou/jurl, Version: " + version)
		fmt.Printf("Usage:\n  %s [options] \"JSON Parser\" URLs\n\nOptions:\n", os.Args[0])
		gnuflag.PrintDefaults()
	}

	gnuflag.Parse(false)
	Args := gnuflag.Args()

	var caCertPool *x509.CertPool
	if ca != "" {
		caCert, err := ioutil.ReadFile(ca)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	if cert != "" && key == "" {
		// Just in case the cert and key are in the same file
		key = cert
	}
	var keypair tls.Certificate
	if cert != "" && key != "" {
		var err error
		keypair, err = tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(Args) < 2 {
		gnuflag.Usage()
		os.Exit(1)
		return
	}

	JQString := Args[0]
	Args = Args[1:]
	var dat map[string]interface{}
	var cacheFiles = make([]string, len(Args))
	var urls = make([](*url.URL), len(Args))

	for i, Arg := range Args {
		u, err := url.Parse(Arg)
		if err != nil {
			fmt.Println("Malformed URL:", err)
			os.Exit(1)
		}
		urls[i] = u
	}

	for i, Arg := range Args {
		h := sha1.New()
		h.Write([]byte(Arg))
		h.Write([]byte(fmt.Sprintf("%d", os.Getuid())))
		bs := h.Sum(nil)

		cacheFile := fmt.Sprintf("%s/jurl_%x", cacheDir, bs)
		cacheFiles[i] = cacheFile

		stat, err := os.Stat(cacheFile)
		if err == nil && !flush && useCache && time.Now().Add(maxAge).After(stat.ModTime()) {
			if debug {
				log.Println("found cache", cacheFile)
			}
			byt, err := ioutil.ReadFile(cacheFile)
			if err == nil {
				if debug {
					log.Println("using cache", cacheFile)
				}
				if includeHeader {
					fmt.Fprintf(os.Stderr, "Header skipped as cache used\nURL: %s\nFile: %s\n", urls[i], cacheFile)
				}
				json.Unmarshal(byt, &dat)
				break
			}
		}
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: certIgnore,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{keypair},
	}
	//http.DefaultTransport.IdleConnTimeout = 10 * time.Second
	client := &http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if followRedirects == false {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	for j := 0; j < maxTries && len(dat) == 0; j++ {
		i := j % len(Args)
		if debug {
			log.Println("HTTP", method, urls[i])
		}
		var err error
		var resp *http.Response
		var req *http.Request
		switch method {
		case "GET", "POST":
			var rdr io.Reader
			if method == "POST" {
				if len(postData) > 0 && postData[0] == '@' {
					f, err := os.Open(postData[1:])
					if err != nil {
						log.Fatal("Unable to open", postData[1:])
					}
					defer f.Close()
					rdr = f
				} else {
					rdr = strings.NewReader(postData)
				}
			}
			req, err = http.NewRequest(method, urls[i].String(), rdr)
			for key, val := range Headers {
				if debug {
					fmt.Printf("Header-- %s: %s\n", key, val)
				}
				req.Header.Set(key, val)
			}
			if err != nil {
				log.Fatal("New request error:", err)
			}
			resp, err = client.Do(req)
		default:
			log.Fatal("Unknown method", method)
		}
		if includeHeader {
			fmt.Fprintf(os.Stderr, "%s %s\n", resp.Proto, resp.Status)
			for key, vals := range resp.Header {
				for _, val := range vals {
					fmt.Fprintf(os.Stderr, "%s: %s\n", key, val)
				}
			}
			fmt.Fprintf(os.Stderr, "\n")
		}
		if err == nil {
			byt, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil {
				err = json.Unmarshal(byt, &dat)
				if err != nil && debug {
					log.Println("cannot unmarshall url:", urls[i], "err:", err)
				}
				if err == nil {
					if !useCache {
						break
					}
					if debug {
						log.Println("writing out file")
					}
					err = ioutil.WriteFile(cacheFiles[i], byt, 0666)
					if err != nil && debug {
						log.Println("error writing file:", err)
					}
					break
				}
			}
		}

		if i%len(Args) == len(Args)-1 {
			time.Sleep(delay)
		}
	}

	query, err := gojq.Parse(JQString)
	if err != nil {
		log.Fatalln(err)
	}
	iter := query.Run(dat) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		if debug {
			fmt.Printf("%#v\n", v)
		} else {
			if raw {
				fmt.Printf("%v\n", v)
			} else {
				out, _ := json.Marshal(v)
				fmt.Println(string(out))
			}
		}
	}
}
