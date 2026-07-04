package coverage

import (
	"testing"
)

func TestParseGoRange_Valid(t *testing.T) {
	fileKey, startLine, endLine, ok := parseGoRange("module/pkg/file.go:1.0,3.0")
	if !ok {
		t.Fatal("parseGoRange should succeed")
	}
	if fileKey != "module/pkg/file.go" {
		t.Errorf("fileKey = %q", fileKey)
	}
	if startLine != 1 || endLine != 3 {
		t.Errorf("lines = %d,%d", startLine, endLine)
	}
}

func TestParseGoRange_Invalid(t *testing.T) {
	tests := []string{
		"",                    // empty
		"foo",                 // no colon
		"foo:bar",             // no comma
		"foo:1.0",             // missing end
		"foo:a.0,3.0",         // non-numeric start
		"foo:1.0,b.0",         // non-numeric end
	}
	for _, tc := range tests {
		_, _, _, ok := parseGoRange(tc)
		if ok {
			t.Errorf("parseGoRange(%q) should fail", tc)
		}
	}
}

func TestParseGoCoverLine_Valid(t *testing.T) {
	bare, key, sl, el, count, ok := parseGoCoverLine("module/pkg/file.go:1.0,3.0 1 1")
	if !ok {
		t.Fatal("parseGoCoverLine should succeed")
	}
	if bare != "file.go" {
		t.Errorf("bare = %q", bare)
	}
	if key != "module/pkg/file.go" {
		t.Errorf("key = %q", key)
	}
	if sl != 1 || el != 3 || count != 1 {
		t.Errorf("got %d,%d,%d", sl, el, count)
	}
}

func TestParseGoCoverLine_ModeLine(t *testing.T) {
	_, _, _, _, _, ok := parseGoCoverLine("mode: set")
	if ok {
		t.Error("mode line should be skipped")
	}
}

func TestParseGoCoverLine_Empty(t *testing.T) {
	_, _, _, _, _, ok := parseGoCoverLine("")
	if ok {
		t.Error("empty line should be skipped")
	}
}

func TestParseGoCoverLine_TooFewFields(t *testing.T) {
	_, _, _, _, _, ok := parseGoCoverLine("file.go:1.0,3.0")
	if ok {
		t.Error("too few fields should fail")
	}
}
