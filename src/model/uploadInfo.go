package model

// UpLoadInfo 小程序发布信息
type UpLoadInfo struct {
	Size struct {
		Total    int `json:"total"`
		Packages []struct {
			Name string `json:"name"`
			Size int    `json:"size"`
		} `json:"packages"`
	} `json:"size"`
}
