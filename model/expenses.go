package model

type Expenses struct {
	ID         string `json:"uuid"`
	Name       string `json:"name"`
	Planned    int    `json:"planned"`
	Spent      int    `json:"spent"`
	Difference int    `json:"difference"`
}
