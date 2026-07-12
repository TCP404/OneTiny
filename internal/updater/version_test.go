package updater

import "testing"

func TestCompareVersionsTreatsVPrefixAsEquivalent(t *testing.T) {
	got, ok := CompareVersions("v1.2.3", "1.2.3")
	if !ok {
		t.Fatal("CompareVersions returned unknown for valid versions")
	}
	if got != 0 {
		t.Fatalf("CompareVersions = %d, want 0", got)
	}
}

func TestCompareVersionsReportsLeftNewer(t *testing.T) {
	got, ok := CompareVersions("1.3.0", "1.2.9")
	if !ok {
		t.Fatal("CompareVersions returned unknown for valid versions")
	}
	if got != 1 {
		t.Fatalf("CompareVersions = %d, want 1", got)
	}
}

func TestCompareVersionsReportsLeftOlder(t *testing.T) {
	got, ok := CompareVersions("1.2.3", "1.2.4")
	if !ok {
		t.Fatal("CompareVersions returned unknown for valid versions")
	}
	if got != -1 {
		t.Fatalf("CompareVersions = %d, want -1", got)
	}
}

func TestCompareVersionsReportsUnknownForEmptyOrDevVersion(t *testing.T) {
	cases := []string{"", "dev"}
	for _, current := range cases {
		if got, ok := CompareVersions(current, "1.0.0"); ok {
			t.Fatalf("CompareVersions(%q, 1.0.0) = %d, true; want unknown", current, got)
		}
	}
}

func TestIsUpdateAvailableReportsNewerLatest(t *testing.T) {
	got := IsUpdateAvailable("1.2.3", "1.2.4")
	if !got.Known {
		t.Fatalf("IsUpdateAvailable returned unknown: %+v", got)
	}
	if !got.Available {
		t.Fatalf("Available = false, want true: %+v", got)
	}
	if got.Reason != "" {
		t.Fatalf("Reason = %q, want empty", got.Reason)
	}
}

func TestIsUpdateAvailableRejectsUnknownCurrentWithReason(t *testing.T) {
	got := IsUpdateAvailable("dev", "1.2.4")
	if got.Known {
		t.Fatalf("Known = true, want false: %+v", got)
	}
	if got.Available {
		t.Fatalf("Available = true, want false: %+v", got)
	}
	if got.Reason == "" {
		t.Fatalf("Reason is empty, want unknown version reason: %+v", got)
	}
}
