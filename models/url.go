package models

type URL struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func NewURL(url string) *URL {
	return &URL{
		ID:  GenerateID(),
		URL: url,
	}
}
func GenerateID() string {
	// Implement a function to generate unique IDs
	return "unique-id"
}
