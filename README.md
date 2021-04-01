# jqURL - URL and JSON parser tool

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
$ jqurl
jqURL - URL and JSON parser tool, Written by paul (paulschou.com)
Usage:
  jqurl [options] "JSON Parser" URLs

Options:
    --cacert FILE      Use certificate authorities, PEM encoded  (Default="")
-C, --cache            Use local cache to speed up static queries
    --cachedir DIR     Path for cache  (Default="/dev/shm")
-E, --cert FILE        Use client cert in request, PEM encoded  (Default="")
-d, --data STRING      Data to use in POST (use @filename to read from file)  (Default="")
    --debug            Debug / verbose output
    --flush            Force redownload, when using cache
-H, --header 'HEADER: VALUE'  Custom header to pass to server
                         (Default="content-type: application/json")
-i, --include          Include header in output
-k, --insecure         Ignore certificate validation checks
    --key FILE         Key file for client cert, PEM encoded  (Default="")
-L, --location         Follow redirects
    --max-age DURATION  Max age for cache  (Default=4h0m0s)
-m, --max-time DURATION  Timeout per request  (Default=15s)
    --max-tries TRIES  Maximum number of tries  (Default=30)
-o, --output FILE      Write output to <file> instead of stdout  (Default="")
-P, --pretty           Pretty print JSON with indents
-r, --raw-output       Raw output, no quotes for strings
-X, --request METHOD   Method to use for HTTP request (ie: POST/GET)  (Default="GET")
    --retry-delay DURATION  Delay between retries  (Default=7s)
```

Envionment variables available for setting:

- HTTPS_PROXY
- HTTP_PROXY


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
Note that the `-r` removes the quotes.

If you have two or more URLs with the same information and want to use them
as backups:
```
[schou]$ jqurl -C ".title" http{,s}://jsonplaceholder.typicode.com/todos/2
"quis ut nam facilis et officia qui"
```
Note that `-C` encourages caching, re-using the previous request.

This is an example of how to POST data and parse the reply:
```
[schou]$ jqurl -P -XPOST -d $'{"method": "POST"}' . https://jsonplaceholder.typicode.com/posts
{
  "id": 101,
  "method": "POST"
}
[schou]$ jqurl -XPOST -d $'{"method": "POST"}' .id https://jsonplaceholder.typicode.com/posts
101
```

As the `--header` or `-H` option works on all header elements, one can use this to both
set any User-Agent or Cookie elements, such as:
```
$ jqurl -H "Cookie: csrftoken=abcd" -H "User-Agent: Mozilla" . https://jsonplaceholder.typicode.com/todos/1
```

For ease of use, single dashed flags can be combined:
```
$ jqurl -CPXGET . https://jsonplaceholder.typicode.com/todos/1
```
Here the `-C` and `-P` need no arguments, while `-X` takes one `"GET"`.

