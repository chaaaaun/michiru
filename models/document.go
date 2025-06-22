package models

import (
	"encoding/json"
	"time"
)

type AnimeDocument struct {
	Aid              json.Number         `json:"aid"`
	MainTitle        string              `json:"mainTitle"`
	OfficialTitles   map[string][]string `json:"officialTitles,omitempty"`
	ShortTitles      map[string][]string `json:"shortTitles,omitempty"`
	SynonymousTitles map[string][]string `json:"synonymousTitles,omitempty"`
	KanaTitles       map[string][]string `json:"kanaTitles,omitempty"`
	CardTitles       map[string][]string `json:"cardTitles,omitempty"`
}

type MetadataDocument struct {
	Id          string    `json:"id"`
	RetrievedAt time.Time `json:"retrievedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	DumpEntries int64     `json:"dumpEntries"`
	DumpTitles  int64     `json:"dumpTitles"`
}
