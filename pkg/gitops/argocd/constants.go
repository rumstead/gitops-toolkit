package argocd

import "regexp"

var (
	serverReplace = regexp.MustCompile("0.0.0.0:[0-9]+")
)

type clusterArgs string

const (
	clusterArgLabels      clusterArgs = "--label"
	clusterArgAnnotations clusterArgs = "--annotation"
)
