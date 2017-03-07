package enulib

type ReleaseNote struct {
	IssueNumber           uint32 `json:"issueNumber"`
	InternalExternalIssue string `json:"internalExternalIssue"`
	Description           string `json:"description"`
}

var ReleaseNotes = []ReleaseNote{
	{22, "internal", "Add support for Ripple"},
	{94, "internal", "Create new function for server info"},
}
