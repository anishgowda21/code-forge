package git

import "errors"

var ErrRepoNotExists = errors.New("repo does not exist")
var ErrRefNotExists = errors.New("ref does not exist")
var ErrPathNotExists = errors.New("path does not exist")
var ErrRepoDirExists = errors.New("repo already exists")
var ErrNotADirectory = errors.New("not a directory")
var ErrNotAFile = errors.New("not a file")
