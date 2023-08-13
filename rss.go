package indeed

import (
	"encoding/xml"
	"time"
)

type RSSFeed struct {
	XMLName     xml.Name  `xml:"rss"`
	Version     string    `xml:"version,attr"`
	Title       string    `xml:"channel>title"`
	Link        string    `xml:"channel>link"`
	Description string    `xml:"channel>description"`
	Items       []RSSItem `xml:"channel>item"`
}

type RSSItem struct {
	Title       string  `xml:"title,omitempty"`
	Link        string  `xml:"link,omitempty"`
	Description string  `xml:"description,omitempty"`
	Author      string  `xml:"author,omitempty"`
	PubDate     RSSTime `xml:"pubDate,omitempty"`
}

type RSSTime struct {
	time.Time
}

func (t RSSTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(t.Format(time.RFC822), start)
}
