package model

type Task struct {
	BranchName  string
	TagName     string
	User        string
	Message     string
	IsProd      bool
	GitRepoURL  string
	ProjectName string
}
