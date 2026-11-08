package state

import (
	"fmt"
	"strings"
)

// Work breakdown groupings, in the order b cycles them on the Day summary. Their values are
// the daemon's distribution groups.
var WorkGroups = []string{"repo", "stream", "issue", "tag", "segment_type"}

// WorkCrumb is one step of a drill into the Day summary's work breakdown: the row chosen and
// the grouping it came from.
type WorkCrumb struct {
	Group string
	ID    int64
	// Tag is the tag name for a tag crumb, which has no id.
	Tag   string
	Label string
}

// WorkRow is one breakdown row as input sees it: its daemon key and label.
type WorkRow struct {
	Key, Label string
}

// NextWorkGroup steps through WorkGroups.
func NextWorkGroup(current string, dir int) string {
	if current == "" {
		current = WorkGroups[0]
	}
	for i, group := range WorkGroups {
		if group == current {
			return WorkGroups[(i+dir+len(WorkGroups))%len(WorkGroups)]
		}
	}
	return WorkGroups[0]
}

// WorkLevelGroup is the grouping shown at the end of path: the top grouping, or the level
// under the last crumb (a repo's streams, a stream's issues).
func WorkLevelGroup(top string, path []WorkCrumb) string {
	if len(path) == 0 {
		if top == "" {
			return WorkGroups[0]
		}
		return top
	}
	if path[len(path)-1].Group == "repo" {
		return "stream"
	}
	return "issue"
}

// WorkDrillable reports whether rows of group open a level beneath them: a repo its
// streams, a stream or a tag its issues. Only these levels take B and enter.
func WorkDrillable(group string) bool {
	return group == "repo" || group == "stream" || group == "tag"
}

// WorkCrumbFor turns a row of group into a crumb, reading the id from keys like "repo:12".
func WorkCrumbFor(group string, row WorkRow) (WorkCrumb, bool) {
	if group == "tag" {
		name, ok := strings.CutPrefix(row.Key, "tag:")
		return WorkCrumb{Group: group, Tag: name, Label: row.Label}, ok && name != ""
	}
	var id int64
	if _, err := fmt.Sscanf(row.Key, group+":%d", &id); err != nil || id <= 0 {
		return WorkCrumb{}, false
	}
	return WorkCrumb{Group: group, ID: id, Label: row.Label}, true
}

// WorkFilters are the repo, stream and tag a path narrows the breakdown to.
func WorkFilters(path []WorkCrumb) (repoID, streamID *int64, tag string) {
	for _, crumb := range path {
		id := crumb.ID
		switch crumb.Group {
		case "repo":
			repoID = &id
		case "stream":
			streamID = &id
		case "tag":
			tag = crumb.Tag
		}
	}
	return repoID, streamID, tag
}

// WorkKey names one breakdown query, so a late result for another can be dropped.
func WorkKey(date, top string, path []WorkCrumb) string {
	key := date + "|" + WorkLevelGroup(top, path)
	for _, crumb := range path {
		key += fmt.Sprintf("|%s:%d:%s", crumb.Group, crumb.ID, crumb.Tag)
	}
	return key
}
