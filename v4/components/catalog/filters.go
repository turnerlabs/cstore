package catalog

func keepFilesWithVersion(files map[string]File, version string) map[string]File {
	filtered := map[string]File{}

	if len(version) == 0 {
		return files
	}

	for key, file := range files {
		if !file.Missing(version) {
			filtered[key] = file
		}
	}

	return filtered
}

func keepFilesWithPaths(files map[string]File, paths []string) map[string]File {
	filtered := map[string]File{}

	if len(paths) == 0 {
		return files
	}

	for _, path := range paths {
		for key, file := range files {
			if file.Path == path {
				filtered[key] = file
			}
		}
	}

	return filtered
}

func keepFilesWithTags(files map[string]File, tags []string, allTags bool) map[string]File {
	filtered := map[string]File{}

	if len(tags) == 0 {
		return files
	}

	for key, file := range files {
		if file.IsRef && len(file.Tags) == 0 {
			filtered[key] = file
		}
	}

	for key, file := range files {
		if allTags {
			if allTagsMatch(file.Tags, tags) {
				filtered[key] = file
			}
		} else {
			if anyTagMatches(file.Tags, tags) {
				filtered[key] = file
			}
		}
	}

	return filtered
}

func allTagsMatch(tags, targetTags []string) bool {
	for _, t := range targetTags {
		matches := false

		for _, tl := range tags {
			if tl == t {
				matches = true
			}
		}

		if !matches {
			return false
		}
	}
	return true
}

func anyTagMatches(tags []string, targetTags []string) bool {
	for _, lt := range targetTags {
		for _, t := range tags {
			if lt == t {
				return true
			}
		}
	}
	return false
}
