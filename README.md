# Introduction #

Domain, feed&mdash;indeed.  
You get the point; naming is hard.  
This little tool lets you track domain events via RSS.  
It works all right, despite its name.

# Usage #

To run the HTTP server with Docker:

```shell
docker run --rm --publish 8080:8080 ghcr.io/axeljohnsson/indeed:main
```

Next, an example of how to query a domain with `curl`.
Include `q` for as many domains as you like.

![curl example](https://static.johnsson.dev/indeed/usage.png)
