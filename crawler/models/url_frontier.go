package models

import (
	"encoding/json"
)

const (
	URL_FRONTIER_STATUS_NEW     int16 = 0
	URL_FRONTIER_STATUS_CRAWLED int16 = 1
	URL_FRONTIER_STATUS_ERROR   int16 = 2
)

type UrlFrontierMetadata struct {
	CitationNumber string   `json:"citation_number"`
	DecisionDate   string   `json:"decision_date"`
	Title          string   `json:"title"`
	Categories     []string `json:"categories"`
	CaseNumbers    []string `json:"case_numbers"`
}

func (m *UrlFrontierMetadata) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func UrlFrontierMetadataFromJson(j []byte) (*UrlFrontierMetadata, error) {
	var m UrlFrontierMetadata
	err := json.Unmarshal(j, &m)
	return &m, err
}
