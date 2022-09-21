package argocd

import "regexp"

var (
	serverReplace = regexp.MustCompile("0.0.0.0:[0-9]+")
)
