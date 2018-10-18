package path

import (
	"fmt"
	"strings"
)

// Format ...
func Format(path string) string {
	const totalLength int = 30
	const ellipsis string = "..."

	pathLength := len(path)

	if pathLength > totalLength {
		trimLength := pathLength - (totalLength - len(ellipsis))

		path = fmt.Sprintf("%s%s", ellipsis, path[trimLength:pathLength])
	} else {
		for i := 0; i < (totalLength - pathLength); i++ {
			path = fmt.Sprintf("%s ", path)
		}
	}

	return path
}

// RemoveFileName ...
func RemoveFileName(path string) string {
	pathEnd := strings.LastIndex(path, "/")
	if pathEnd == -1 {
		return ""
	}
	return path[:pathEnd+1]
}

// BuildPath ...
func BuildPath(root, path string) string {
	if len(root) > 0 {
		if strings.LastIndex(root, "/") != len(root)-1 {
			return fmt.Sprintf("%s/%s", root, path)
		}
		return fmt.Sprintf("%s%s", root, path)
	}
	return path
}
