package action

import "testing"

func TestBuildMetaEventKey(t *testing.T) {
	if got := BuildMetaEventKey("51", "AppLaunch"); got != "51_AppLaunch" {
		t.Fatalf("BuildMetaEventKey = %q", got)
	}
}

func TestBuildMetaAttrRelationKey(t *testing.T) {
	if got := BuildMetaAttrRelationKey("51", "AppLaunch", "xwl_browser"); got != "51_AppLaunch_xwl_browser" {
		t.Fatalf("BuildMetaAttrRelationKey = %q", got)
	}
}

func TestBuildAttributeKey(t *testing.T) {
	if got := BuildAttributeKey("51", 2, "xwl_browser"); got != "51_xwl_2_xwl_browser" {
		t.Fatalf("BuildAttributeKey = %q", got)
	}
}
