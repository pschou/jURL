package main

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/itchyny/gojq"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var version = "debug"

var debug = false

func main() {
	var maxTries int
	var delay, maxAge time.Duration
	var debug, raw, includeHeader, certIgnore, flush bool
	var cert, key, ca, cache string
	flag.BoolVar(&flush, "flush", false, "Force download, don't use cache.")
	flag.BoolVar(&debug, "debug", false, "Debug / verbose output")
	flag.IntVar(&maxTries, "maxtries", 30, "Maximum number of tries")
	flag.BoolVar(&raw, "r", false, "Raw output, no quotes for strings")
	flag.BoolVar(&includeHeader, "i", false, "Include header in output")
	flag.BoolVar(&certIgnore, "k", false, "Ignore certificate validation checks")
	flag.StringVar(&ca, "ca", "", "Use certificate authorities, PEM encoded")
	flag.StringVar(&cert, "cert", "", "Use client cert in request, PEM encoded")
	flag.StringVar(&key, "certkey", "", "Key file for client cert, PEM encoded")
	flag.StringVar(&cache, "cache", "/dev/shm", "Path for cache")
	flag.DurationVar(&delay, "delay", 7*time.Second, "Delay between retries")
	flag.DurationVar(&maxAge, "maxage", 4*time.Hour, "Max age for cache")

	flag.Usage = func() {
		fmt.Println("Simple JSON URL downloader and parser tool, Written by paul (paulschou.com), Docs: github.com/pschou/jurl, Version: " + version)
		fmt.Printf("Usage:\n  %s [options] \"JQuery\" URLs\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	Args := flag.Args()

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
		flag.Usage()
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

		cacheFile := fmt.Sprintf("%s/jurl_%x", cache, bs)
		cacheFiles[i] = cacheFile

		stat, err := os.Stat(cacheFile)
		if err == nil && !flush && time.Now().Add(maxAge).After(stat.ModTime()) {
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

	for j := 0; j < maxTries && len(dat) == 0; j++ {
		i := j % len(Args)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: certIgnore,
			RootCAs:            caCertPool,
			Certificates:       []tls.Certificate{keypair},
		}
		if debug {
			log.Println("http get", urls[i])
		}
		resp, err := http.Get(urls[i].String())
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
				fmt.Printf("%s\n", v)
			} else {
				out, _ := json.Marshal(v)
				fmt.Println(string(out))
			}
		}
	}
}
