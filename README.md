# JURL - JSON URL downloader and parser

Imagine how many times from the command line shell, such as BaSH, one calls `curl` and then pipes that output to `jq`.  What's even worse is that the same rest endpoint is used multiple times in one script or between scripts.  Enter `jurl`, this compact tool gives the basic set features of `curl` and `jq` in one binary.  

For efficiency, `jurl` will store a cached responce in `/dev/shm/jurl_...`.
This way multiple queries to the same endpoint will return the same results and
execute faster.

## Why do I care?



## Syntax

```
$ ./jurl
Simple JSON URL download and parser tool, Written by paul (paulschou.com), Docs: github.com/pschou/jurl, Version: 0.1.20210325.1930

Syntax: ./jurl "JQ_QUERY" [URL]
```

## Example

Here is an example showing usage with a rest endpoint
```
[schou]$ curl https://jsonplaceholder.typicode.com/todos/1
{
  "userId": 1,
  "id": 1,
  "title": "delectus aut autem",
  "completed": false
}
[schou]$ jurl .title https://jsonplaceholder.typicode.com/todos/1
delectus aut autem
```

