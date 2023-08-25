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

```shell
curl --silent 'http://localhost:8080/feed?q=example.com' | xmllint --format -
```

You should see something like this:

```xml
<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Domain Events</title>
    <link>/feed?q=example.com</link>
    <description>Domain events for: example.com.</description>
    <item>
      <link>https://rdap.org/domain/EXAMPLE.COM</link>
      <description>example.com: expiration</description>
      <guid>0e7b8746deb1b3df50b53bd3fa1df6f795e130088f3dbee4fbcd559b99ea7e46</guid>
      <pubDate>13 Aug 24 04:00 UTC</pubDate>
    </item>
    <item>
      <link>https://rdap.org/domain/EXAMPLE.COM</link>
      <description>example.com: last update of RDAP database</description>
      <guid>f1194c798bf1a1a603735c0ca0b536f59835c8ded794f215410b2192fe7677c7</guid>
      <pubDate>25 Aug 23 18:30 UTC</pubDate>
    </item>
    <item>
      <link>https://rdap.org/domain/EXAMPLE.COM</link>
      <description>example.com: last changed</description>
      <guid>264aaecf302ed10f175731ded269a76e2ac202212ac70cf6e73977e6ba033f5b</guid>
      <pubDate>14 Aug 23 07:01 UTC</pubDate>
    </item>
    <item>
      <link>https://rdap.org/domain/EXAMPLE.COM</link>
      <description>example.com: registration</description>
      <guid>8c0e7bcead41a573c598c2ab9ae7e95fde486b0d7307b115a1da9b6d6fbb8c4a</guid>
      <pubDate>14 Aug 95 04:00 UTC</pubDate>
    </item>
  </channel>
</rss>
```
