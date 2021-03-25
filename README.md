# JURL - JSON URL downloader and parser

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
[schou]$ ./jurl .title https://jsonplaceholder.typicode.com/todos/1
delectus aut autem
```
