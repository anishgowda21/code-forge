package api

import (
	"regexp"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var repoNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
