package model

// PublishPostData 自动发布系统POSTData
type PublishPostData struct {
	Branch  string `json:"branch"`
	TagName string `json:"tagName"`
	User    string `json:"user"`
	Message string `json:"message"`
	GitRepo string `json:"gitRepo"`
	Name    string `json:"name"`
	IsProd  bool   `json:"isProd"`
}
