# Simple Statik

A simple static server. It takes a simple configuration
at runtime to serve static files or contents.

Urls are matched as an exact match by default, but can be 
changed to behave as a prefix with a `prefix=true` config.

## Running

By default will run on port `8081`.

```
$ cat example.config
/
-> message={"hello": "world"}
-> header="Content-Type=application/json"

# Endpoint for healthcheck.
/healthz
-> message={"alive": true}
-> header="Content-Type=application/json"

# This is a catch all for any route that wasn't matched.
_
-> message=No route found
-> http_status_code=404

$ simple-statik -config=example.config &

$ curl -i localhost:8081/
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 27 Mar 2019 00:06:20 GMT
Content-Length: 18

{"hello": "world"}

$ curl -i localhost:8081/healthz
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 27 Mar 2019 00:06:23 GMT
Content-Length: 15

{"alive": true}

$ curl -i "localhost:8081/not-exist"
HTTP/1.1 404 Not Found
Date: Wed, 27 Mar 2019 00:06:28 GMT
Content-Length: 14
Content-Type: text/plain; charset=utf-8

No route found
```

## Configuration

```
<url_exact_match>
-> <key>=<value>
-> <key>=<value>
```

A `<url_exact_match>` can be any url path, or `_` which 
behaves as a catch all if you customise what to do if no
paths are matched. Otherwise a `404` is returned.

Possible keys:

- file: static file to serve
- folder: folder to serve static file contents from
- message: a static message to serve
- http_status_code: http status code to return in response
- header: http header to return in response
- prefix: should this url be treated as a prefix(rather than an exact match)

### Return JSON contents

```
/json
-> message={"message": "hello world", "version": "v1"}
-> header="Content-Type=application/json"

/json/v2
-> prefix=true
-> message={"message": "other world", "version": "v2"}
-> header="Content-Type=application/json"
```

This would match url `/json`, `/json/v2` and `/json/v2/anything`.
This would not match `json/`.

### Return file contents

```
/
-> file=index.html

/static
-> prefix=true
-> folder=/path/to/static/
```

This will return a static file `index.html` for the path
`/`. `/static/site.js` will try and read the file `/path/to/static/site.js`
and return the file in the response.

## Example Configuration

```
# url -> static content or status code
/ 
-> file=index.html

/iceme/privacy_policy /iceme/privacy-policy
-> file=iceme_privacy_policy.html # Serve from file

/static
-> prefix=true 
-> folder=static/ # Serve from folder

/json 
-> message={"hello": "world", "version": "v1"} 
-> header=Content-Type=application/json

/json/v2
-> message={"hello": "world", "version": "v2"} 
-> header="Content-Type=application/json"

/404 
-> http_status_code=404 
-> message=This is a 404

/500 
-> http_status_code=500
-> message="Server error"

# '_' is a catch all for anything that doesn't match the above
_ 
-> http_status_code=404
-> message="No route found"
```
