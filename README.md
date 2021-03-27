# JURL - JSON URL downloader and parser

Imagine how many times the command-line shell, such as BaSH, calls `cURL` and
then pipes that output to `jq`.  What's even worse is that the same rest
endpoint is used multiple times in one script or between scripts.  Enter
`jurl`, this compact tool gives the basic set features of `cURL` and `jq` in
one binary.

This way multiple queries to the same endpoint will return the same results and
execute faster.

## Why do I care?

- Retries until JSON is returned
- Built in JQ to select the right JSON element
- Prints out a simple string if an element is selected, or child JSON
- Keeps a cache to avoid overloading the backend rest endpoint
- Open source and free

For efficiency, `jurl` will store a cached response in `/dev/shm/jurl_...`.
This cache will be used before any of the provided URLs are downloaded.
For example, OpenStack or CS2 infrastructure both of which provide metadata / JSON
endpoints for collecting system details.  These details don't change, but they
can be queried multiple times for many uses, such as metrics, system
identification, and health monitoring.  This tool is the ideal script driven choice.

This binary is a portable package,
statically compiled binary, and the minimalist replies mean it suits well to
sit inside any script and run in a shell escape.


## Syntax

```
$ ./jurl
Simple JSON URL download and parser tool, Written by paul (paulschou.com)

Usage:
  ./jurl [options] "JSON Parser" URLs

Options:
--cacert (= "")
    Use certificate authorities, PEM encoded
--cache (= "/dev/shm")
    Path for cache
--cert (= "")
    Use client cert in request, PEM encoded
-d (= "")
    Data to use in POST (use @filename to read from file)
--debug  (= false)
    Debug / verbose output
--delay  (= 7s)
    Delay between retries
--flush  (= false)
    Force download, don't use cache.
-i  (= false)
    Include header in output
-k  (= false)
    Ignore certificate validation checks
--key (= "")
    Key file for client cert, PEM encoded
--maxage  (= 4h0m0s)
    Max age for cache
--maxtries  (= 30)
    Maximum number of tries
-r  (= false)
    Raw output, no quotes for strings
-x (= "GET")
    Method to use for HTTP request (ie: POST/GET)
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
instead of asking `cURL` to blindly fetch something, we use `jURL` which does
this task and ensures success:

```
[schou]$ jurl -r .title https://jsonplaceholder.typicode.com/todos/1
delectus aut autem
```

If you have two or more URLs with the same information and want to use them
as backups:
```
[schou]$ jurl ".title" http{,s}://jsonplaceholder.typicode.com/todos/2
"quis ut nam facilis et officia qui"
```

