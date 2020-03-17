package model

type TinyLogo struct {
	Size Size `json:"size"`
}
type Packages struct {
	Name string  `json:"name"`
	Size float64 `json:"size"`
}
type Size struct {
	Total    float64    `json:"total"`
	Packages []Packages `json:"packages"`
}
