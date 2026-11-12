package git

import "fmt"

func getTreeObjectType(mode, object string) (RefTreeObjectType, error) {
	//Ref: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
	switch object {
	case "tree":
		return TypeDir, nil
	case "commit":
		return TypeSubmodule, nil
	case "blob":
		if mode == "120000" {
			return TypeSymlink, nil
		}
		return TypeFile, nil
	default:
		return "", fmt.Errorf("get tree object type: unexpected object:%q", object)
	}
}
