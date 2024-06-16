package ui

type DocumentsIndexPage struct {
	Error     string
	Query     string
	Documents []SearchResult
	Answer    string
}

type SearchResult struct {
	Title           string
	DocumentID      string
	Chunk           string
	Content         string
	Page            string
	Score           string
	AcceptableScore bool
}
