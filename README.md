# ReqStatSrv

An http server to for **fake responses with weird
behaviours**: delayed responses, connection cuts,
slow responses...

## Usage:

```bash
docker pull dhontecillas/reqstatsrv:v0.4
docker run --rm -p 9876:9876 \
    -v ./example:/etc/reqstatsrv \
    dhontecillas/reqstatsrv:v0.4 \
    /reqstatsrv /etc/config/example.json
```

or from the base repo run:

```
make dockerrun
```

## Example: 

# Configuration

The server is be configured using **JSON** config file as the only
argument.

At the root level, you can configure:
- `port`: the binding port
- `host`: the address to bind to 
- `endpoints`: the list of endpoints to serve


## Endpoint

The endpoint has:

- `method`: an uppercased http method: 'GET', 'POST', ...
- `path_pattern`: The server uses the Go's standard lib router, so, 
    to define a route you can follow the 
    [patterns documentation](https://pkg.go.dev/net/http#hdr-Patterns-ServeMux),
    with the slight difference that the method is specified in the
    `method` field.
- `behaviour`: configuration about how we want the response to
    behave (see explanation below).
- `content`: configuration about the content to serve with the
    request (**headers** and **body** are generated at this level).

## Behaviours

### Delay

Waits for a certain amount of time defined by a distribution
of milliseconds ramp.

You define a list of key (milliseconds) / value (probability) pairs.
Those are the probability of having a dalay below that "key" milliseconds.

Example:

| Key | Val  |
|-----|------|
| 0   | 0.05 |
| 50  | 0.20 |
| 150 | 0.0  |
| 300 | 0.25 |
| 500 | 0.5  |


With the above values, 5%  of request will have no delay at all, a 20% will
be between 0 and 50 ms of delay, there will be no requests with a dealay between
50 and 150 ms, and then there will be a 25% of requests with a delay between 150
and 300ms, and the rest of requests will have a delay between 300ms and 500ms.

```json
{
    "name": "delayer",
    "config": {
        "delay_millis_distribution": [
            {"key": 0, "val": 0.05},
            {"key": 10, "val": 0.2},
            {"key": 50, "val": 0.0},
            {"key": 100, "val": 0.5},
            {"key": 200, "val": 0.0},
            {"key": 700, "val": 0.25}
        ],
        "seed": 1
    }
}
```

### Slower

Makes any content to be served at a given rate. The flush call can
be also tweaked.

```json
{
    "name": "slower",
    "config": {
        "max_bytes_per_second": 100,
        "flush_bytes": 2
    }
},
```

### Status Distributor

The status code to produce is stored in the incoming request 
context, so, the content providers can read it (and "respect" the 
choice taken at the behaviour level), and might be able to select
the content based on the status code.

Here we define a similar distribution as in the delay part, but
with the difference that the produced values will not be interpolated.

| Key | Val |
|-----|-----|
| 200 | 0.5 |
| 201 | 0.3 |
| 204 | 0.2 |

So, with that config, we will have `200` in 50% of the requests, `201` in
30% of the requests and a `204` in the rest.

Values are normalized, meaning that we could set the values to 5, 3 and 2
and the result would be the same.

```json
{
    "name": "status_distributor",
    "config": {
        "code_distribution": [
            {"key": 200, "val": 0.5},
            {"key": 201, "val": 0.3},
            {"key": 204, "val": 0.2}
        ],
        "seed": 1
    }
}
```

### Connection Closer

This behaviour will close the connection with the frequency selected: the 
range for freq must be between 0.0 and 1.0.

```json
{
    "name": "connection_closer",
    "config": {
        "freq": 0.5,
        "seed": 1
    }
}
```

## Content

### Empty

Will just set the `Content-Length` to 0, and return without writing the body.

### Dummy

This content is here for testing purposes, and will surely be deleted on the
1.0 version.

### Directory

Serve the files that are set in the directory path. The Content-Type for each
file will be set according to the file extension. 

With the `attempt_extensions` option, we can provide a list of file extensions
to attempt when we cannot match directly the path to a file in the filesystem.
Useful for example if we want to have a path like `/foo/bar/` to return a 
json file (because the content-type is infered from the file extension), if 
we add the `.json` file, it will try to find the `/foo/bar.json` file. This
also allows to have a `/foo/bar/` dir in the filesystem, and put inside files
like `/foo/bar/1`, `/foo/bar/2`, and have those inside the directory.

With the `dunder_querystrings` option (`true` / `false`), sorts the query params alphabetically, and
joins them after the file name using "dunder" (double underscore) separators
between key, value pairs. Key Value pairs are split by using a single
underscore. So a path like: /foo/bar?a=foo&b=bar will become:
/foo/bar__a_foo__b_bar.

```json
{
    "source": "directory",
    "config": {
        "dir": "./example/data",
        "attempt_extensions": ["json", "yaml"],
        "dunder_querystrings": true
    }
}
```

### File

The response for the endpoint is directly mapped to the given file.
The content-type is extracted from the file extension.

```json
{
    "method": "GET",
    "path_pattern": "/file",
    "behaviour": [],
    "content": {
        "source": "file",
        "config": {
            "path": "./example/data/bar.json" 
        }
    }
}
```

### Stats

This content reports basic stats for the served requests.

Currently it only shows "instant" requests count (the number
of requests served for each endpoint in the past second).

```json
"content": {
    "source": "stats",
    "config": {}
}
```

### Proxy

This content forwards the received request to another server 
to retrieve the contents, allowing to apply behaviours to some
existing endpoints (delay, slow sending of bytes, or change the
status code).

```json
"content": {
    "source": "proxy",
    "config": {
        "proxy_url": "http://127.0.0.1:9877"
    }
}
```

### Status Content Selector 

This content allows to specify child contents depending on the 
status code that the behaviour layer selected. This allows to 
forward a request to the proxy if it is a successful status code,
or fake it with some local file (or use empty response), for other
status codes.

```json
"content": {
    "source": "proxy",
    "config": {
        "default_content" : { 
            "source": "proxy",
            "config": {
                "proxy_url": "http://127.0.0.1:9999"
            }
        },
        "status_contents": [
            "from": 500,
            "to": 600,
            "content": {
                "source": "file",
                "config": {
                    "path": "./example/data/internal_server_error.json" 
                }
            }
        ]
    }
}
```
