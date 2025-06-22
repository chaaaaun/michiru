package handlers

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	"michiru/models"
)

func ParseDump(b []byte) ([]models.AnimeDocument, *models.MetadataDocument, error) {
	var meta models.MetadataDocument
	meta.RetrievedAt = time.Now().UTC().Truncate(time.Second)

	anime := make([]models.AnimeXMLItem, 0)
	r := bytes.NewReader(b)
	d := xml.NewDecoder(r)
	for {
		token, err := d.Token()
		if token == nil || err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "anime" {
				var a models.AnimeXMLItem
				if err := d.DecodeElement(&a, &t); err != nil {
					log.Printf("invalid formatting for XML item: %s", err)
					continue
				}
				anime = append(anime, a)
			}
		case xml.Comment:
			re, err := regexp.Compile(`: (.*) \((\d*)\D*(\d*)`)
			if err != nil {
				return nil, nil, err
			}

			groups := re.FindSubmatch(t)

			datetime, _ := time.Parse("Mon Jan 2 15:04:05 2006", string(groups[1]))
			meta.UpdatedAt = datetime.UTC()
			entries, _ := strconv.ParseInt(string(groups[2]), 10, 0)
			meta.DumpEntries = entries
			titles, _ := strconv.ParseInt(string(groups[3]), 10, 0)
			meta.DumpTitles = titles
		}
	}
	coll := make([]models.AnimeDocument, 0)
	for _, item := range anime {
		coll = append(coll, item.ToDocument())
	}

	return coll, &meta, nil
}
