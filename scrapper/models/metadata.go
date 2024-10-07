package models

var EmptyMetadata Metadata

type Metadata struct {
	Id                 string `json:"id"`
	Title              string `json:"title"`
	Defendant          string `json:"defendant"`
	Number             string `json:"number"`
	CitationNumber     string `json:"citation_number"`
	Classification     string `json:"classification"`
	Year               string `json:"year"`
	JudicalInstitution string `json:"judicial_institution"`
	Judges             string `json:"judges"`
	Verdict            string `json:"verdict"`
	VerdictMarkdown    string `json:"verdict_markdown"`
	DecisionDate       string `json:"decision_date"`
	PdfUrl             string `json:"pdf_url"`
}
