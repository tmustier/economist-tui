package rss

import "sort"

// SectionInfo describes a canonical Economist section and its aliases.
type SectionInfo struct {
	Primary string
	Path    string
	Aliases []string
}

// SectionList returns the canonical section list with primary aliases sorted by path.
func SectionList() []SectionInfo {
	pathToAliases := make(map[string][]string)
	for alias, path := range Sections {
		pathToAliases[path] = append(pathToAliases[path], alias)
	}

	paths := make([]string, 0, len(pathToAliases))
	for path := range pathToAliases {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	sections := make([]SectionInfo, 0, len(paths))
	for _, path := range paths {
		aliases := pathToAliases[path]
		sort.Strings(aliases)
		primary := shortestString(aliases)
		sections = append(sections, SectionInfo{Primary: primary, Path: path, Aliases: aliases})
	}

	return sections
}

func shortestString(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	shortest := strs[0]
	for _, s := range strs[1:] {
		if len(s) < len(shortest) {
			shortest = s
		}
	}
	return shortest
}
