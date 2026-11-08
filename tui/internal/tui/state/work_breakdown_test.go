package state

import "testing"

func TestWorkDrillLevelsAndFilters(t *testing.T) {
	if got := WorkLevelGroup("", nil); got != "repo" {
		t.Fatalf("default level = %q", got)
	}
	crumb, ok := WorkCrumbFor("repo", WorkRow{Key: "repo:12", Label: "OSS"})
	if !ok || crumb.ID != 12 {
		t.Fatalf("crumb = %+v %v", crumb, ok)
	}
	path := []WorkCrumb{crumb}
	if got := WorkLevelGroup("repo", path); got != "stream" {
		t.Fatalf("under a repo = %q", got)
	}
	path = append(path, WorkCrumb{Group: "stream", ID: 4})
	if got := WorkLevelGroup("repo", path); got != "issue" {
		t.Fatalf("under a stream = %q", got)
	}
	repo, stream, _ := WorkFilters(path)
	if repo == nil || *repo != 12 || stream == nil || *stream != 4 {
		t.Fatalf("filters = %v %v", repo, stream)
	}
	tag, ok := WorkCrumbFor("tag", WorkRow{Key: "tag:urgent", Label: "+urgent"})
	if !ok || tag.Tag != "urgent" || WorkLevelGroup("tag", []WorkCrumb{tag}) != "issue" {
		t.Fatalf("a tag should open its issues, got %+v", tag)
	}
	if _, _, name := WorkFilters([]WorkCrumb{tag}); name != "urgent" {
		t.Fatalf("tag filter = %q", name)
	}
	if WorkDrillable("segment_type") || WorkDrillable("issue") {
		t.Fatalf("segments and issues open nothing")
	}
	if NextWorkGroup("segment_type", 1) != "repo" || NextWorkGroup("repo", -1) != "segment_type" {
		t.Fatalf("cycle should wrap")
	}
}
