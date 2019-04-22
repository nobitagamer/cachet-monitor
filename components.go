package cachet

import "encoding/json"

func UnmarshalComponents(data []byte) (Components, error) {
	var r Components
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Components) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Components struct {
	Meta Meta    `json:"meta"`
	Data []Datum `json:"data"`
}

type Datum struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Link        string      `json:"link"`
	Status      int64       `json:"status"`
	Order       int64       `json:"order"`
	GroupID     int64       `json:"group_id"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	DeletedAt   interface{} `json:"deleted_at"`
	Enabled     bool        `json:"enabled"`
	StatusName  string      `json:"status_name"`
	Tags        Tags        `json:"tags"`
}

type Tags struct {
	Empty *string `json:",omitempty"`
	Cpi   *string `json:"cpi,omitempty"`
}

type Meta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Total       int64 `json:"total"`
	Count       int64 `json:"count"`
	PerPage     int64 `json:"per_page"`
	CurrentPage int64 `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
	Links       Links `json:"links"`
}

type Links struct {
	NextPage     interface{} `json:"next_page"`
	PreviousPage interface{} `json:"previous_page"`
}
