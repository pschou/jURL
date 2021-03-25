package main

import (
	//"crypto/tls"
	//"errors"
	"fmt"
	//"github.com/miekg/dns"
	//"math/rand"
	"crypto/sha1"
	"encoding/json"
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
	if len(os.Args) > 1 && os.Args[1] == "-debug" {
		debug = true
		os.Args = os.Args[1:]
	}

	if len(os.Args) < 3 {
		fmt.Println("Simple JSON URL download and parser tool, Written by paul (paulschou.com), Docs: github.com/pschou/jurl, Version: "+version+
			"\n\nSyntax:", os.Args[0], "\"JQ_QUERY\" [URL]")
		os.Exit(1)
		return
	}

	h := sha1.New()
	h.Write([]byte(os.Args[2]))
	h.Write([]byte(fmt.Sprintf("%d", os.Getuid())))
	bs := h.Sum(nil)

	cacheFile := fmt.Sprintf("/dev/shm/jurl_%x", bs)
	var dat map[string]interface{}

	if _, err := os.Stat(cacheFile); err == nil {
		if debug {
			log.Println("found cache", cacheFile)
		}
		byt, err := ioutil.ReadFile(cacheFile)
		if err == nil {
			if debug {
				log.Println("using cache", cacheFile)
			}
			json.Unmarshal(byt, &dat)
		}
	}

	u, err := url.Parse(os.Args[2])
	if err != nil {
		fmt.Println("Malformed URL:", err)
		os.Exit(1)
	}
	if debug {
		fmt.Println("found", u)
	}

	for len(dat) == 0 {
		if debug {
			log.Println("http get", u)
		}
		resp, err := http.Get(u.String())
		if err == nil {
			byt, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err == nil {
				err = json.Unmarshal(byt, &dat)
				if err != nil && debug {
					log.Println("cannot unmarshall url:", u, "err:", err)
				}
				if err == nil {
					if debug {
						log.Println("writing out file")
					}
					err = ioutil.WriteFile(cacheFile, byt, 0666)
					if err != nil && debug {
						log.Println("error writing file:", err)
					}
					break
				}
			}
		}

		time.Sleep(5 * time.Second)
		switch u.Scheme {
		case "http":
			u.Scheme = "https"
		case "https":
			u.Scheme = "http"
		}
	}

	query, err := gojq.Parse(os.Args[1])
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
			out, _ := json.Marshal(v)
			fmt.Printf(string(out))
		}
	}
}
