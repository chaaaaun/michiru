package models

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
)

// AniDB XML structs

type AnimeXMLItem struct {
	XMLName xml.Name       `xml:"anime"`
	Aid     int            `xml:"aid,attr"`
	Titles  []TitleXMLItem `xml:"title"`
}

type TitleXMLItem struct {
	XMLName  xml.Name `xml:"title"`
	Type     string   `xml:"type,attr"`
	Language string   `xml:"lang,attr"`
	Value    string   `xml:",chardata"`
}

func (anime AnimeXMLItem) ToDocument() AnimeDocument {
	doc := AnimeDocument{
		Aid:              json.Number(strconv.Itoa(anime.Aid)),
		OfficialTitles:   map[string][]string{},
		ShortTitles:      map[string][]string{},
		SynonymousTitles: map[string][]string{},
		KanaTitles:       map[string][]string{},
		CardTitles:       map[string][]string{},
	}
	for _, title := range anime.Titles {
		switch title.Type {
		case "main":
			doc.MainTitle = title.Value
		case "official":
			doc.OfficialTitles[title.Language] = append(
				doc.OfficialTitles[title.Language], title.Value,
			)
		case "syn":
			doc.SynonymousTitles[title.Language] = append(
				doc.SynonymousTitles[title.Language], title.Value,
			)
		case "short":
			doc.ShortTitles[title.Language] = append(
				doc.ShortTitles[title.Language], title.Value,
			)
		case "kana":
			doc.KanaTitles[title.Language] = append(
				doc.KanaTitles[title.Language], title.Value,
			)
		case "card":
			doc.CardTitles[title.Language] = append(
				doc.CardTitles[title.Language], title.Value,
			)
		}
	}

	return doc
}
