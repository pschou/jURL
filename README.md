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
- Open source and free

For efficiency, `jurl` will store a cached response in `/dev/shm/jurl_...`.
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
  ./jurl [options] "JQuery" URLs

Options:
  -ca string
        Use certificate authorities, PEM encoded
  -cache string
        Path for cache (default "/dev/shm")
  -cert string
        Use client cert in request, PEM encoded
  -certkey string
        Key file for client cert, PEM encoded
  -debug
        Debug / verbose output
  -delay duration
        Delay between retries (default 7s)
  -flush
        Force download, don't use cache.
  -i    Include header in output
  -k    Ignore certificate validation checks
  -maxage duration
        Max age for cache (default 4h0m0s)
  -maxtries int
        Maximum number of tries (default 30)
  -r    Raw output, no quotes for strings
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

