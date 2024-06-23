package openalex

type WorkObject struct {
	OpenAccess      OpenAccess `json:"open_access"`
	CitedByCount    uint32     `json:"cited_by_count"`
	Title           string     `json:"title"`
	PrimaryTopic    Topic      `json:"primary_topic"`
	Language        string     `json:"language"`
	PublicationYear uint32     `json:"publication_year"`
	PublicationDate string     `json:"publication_date"`
	Text            string     `json:"text"`
	Id              string     `json:"id"`
}

type OpenAccess struct {
	IsOA     bool   `json:"is_oa"`
	OAURL    string `json:"oa_url"`
	OAStatus string `json:"oa_status"`
}

type Topic struct {
	ID     string `json:"id"`
	Domain Domain `json:"domain"`
}

type Domain struct {
	DisplayName string `json:"display_name"`
}
