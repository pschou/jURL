# jqURL - URL downloader and JSON parser

Imagine how many times the command-line shell, such as BaSH, calls `cURL` and
then pipes that output to `jq`.  What's even worse is that the same rest
endpoint is used multiple times in one script or between scripts.  Enter
`jqurl`, this compact tool gives the basic set features of `cURL` and `jq` in
one binary.

This way multiple queries to the same endpoint will return the same results and
execute faster.

Pronounced like "j curl".

## Why do I care?

- Retries until JSON is returned
- Built in JQ to select the right JSON element
- Prints out a simple string if an element is selected, or child JSON
- Keeps a cache to avoid overloading the backend rest endpoint
- Open source and free

For efficiency, `jqurl` can store a cached response in a temporary directory (like `/dev/shm/jqurl_...`).
This cache will be used in future queries before any of the provided URLs are downloaded (when using the flag -C).
For example, OpenStack or CS2 infrastructure both of which provide metadata / JSON
endpoints for collecting system details.  These details don't change, but they
can be queried multiple times for many uses, such as metrics, system
identification, and health monitoring.  This tool is the ideal script driven choice.

This binary is a portable package,
statically compiled binary, and with the minimalist output, it is tailored to and suits well for usage
inside any script, invoked via a shell command.


## Syntax

```
$ ./jqurl
Simple JSON URL download and parser tool, Written by paul (paulschou.com)
Usage:
  jqurl [options] "JSON Parser" URLs

Options:
-C, --cache
          Use local cache to speed up static queries
-E, --cert FILE (Default= "")
          Use client cert in request, PEM encoded
-H, --header "Key: Value"  (Default= "content-type: application/json")
          Custom header to pass to server
-L, --location
          Follow redirects
-X, --request METHOD (Default= "GET")
          Method to use for HTTP request (ie: POST/GET)
    --cacert FILE (Default= "")
          Use certificate authorities, PEM encoded
    --cachedir DIR (Default= "/dev/shm")
          Path for cache
-d, --data STRING (Default= "")
          Data to use in POST (use @filename to read from file)
    --debug
          Debug / verbose output
    --flush
          Force redownload, when using cache
-i, --include
          Include header in output
-k, --insecure
          Ignore certificate validation checks
    --key FILE (Default= "")
          Key file for client cert, PEM encoded
    --maxage DURATION  (Default= 4h0m0s)
          Max age for cache
    --maxtries TRIES  (Default= 30)
          Maximum number of tries
-r, --raw-output
          Raw output, no quotes for strings
    --retry-delay DURATION  (Default= 7s)
          Delay between retries
```

## What we want

Here is an example showing usage using `curl` on a rest endpoint:
```
[schou]$ curl https://jsonplaceholder.typicode.com/todos/1
{
  "userId": 1,
  "id": 1,
  "title": "delectus aut autem",
  "completed": false
}
```

How one would typically use cURL and JQ together:
```
[schou]$ curl -s https://jsonplaceholder.typicode.com/todos/1 | jq -r .title
delectus aut autem
```

The problem here is if `cURL` hits a broken endpoint or is fed a 404 error
message, JQ parses junk.  The old saying, junk in produces junk out.  So
instead of asking `cURL` to blindly fetch something, we use `jqURL` which does
this task and ensures success:

```
[schou]$ jqurl -Cr .title https://jsonplaceholder.typicode.com/todos/1
delectus aut autem
```

If you have two or more URLs with the same information and want to use them
as backups:
```
[schou]$ jqurl -C ".title" http{,s}://jsonplaceholder.typicode.com/todos/2
"quis ut nam facilis et officia qui"
```

This is an example of how to POST data and parse the reply:
```
[schou]$ jqurl -XPOST -d $'{"method": "POST"}' . https://jsonplaceholder.typicode.com/posts
{"id":101,"method":"POST"}
[schou]$ jqurl -XPOST -d $'{"method": "POST"}' .id https://jsonplaceholder.typicode.com/posts
101
```

