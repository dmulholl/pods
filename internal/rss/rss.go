package rss

import "encoding/xml"

type RSS struct {
	XMLName  xml.Name  `xml:"rss"`
	Channels []Channel `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Items   []Item   `xml:"item"`
}

type Item struct {
	XMLName   xml.Name  `xml:"item"`
	Title     string    `xml:"title"`
	PubDate   string    `xml:"pubDate"`
	GUID      string    `xml:"guid"`
	Enclosure Enclosure `xml:"enclosure"`
	Episode   int       `xml:"episode"`
}

type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Type    string   `xml:"type,attr"`
	Length  int64    `xml:"length,attr"`
}
