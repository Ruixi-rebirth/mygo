package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFprintf(t *testing.T) {
	oldColor := ColorEnabled
	ColorEnabled = false
	defer func() { ColorEnabled = oldColor }()

	tests := []struct {
		level   Level
		format  string
		args    []any
		contain string
	}{
		{Success, "built %s", []any{"main"}, "built main"},
		{Error, "failed: %v", []any{"oops"}, "failed: oops"},
		{Warn, "warning: %s", []any{"deprecated"}, "warning: deprecated"},
		{Info, "running %s", []any{"test"}, "running test"},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		Fprintf(&buf, tt.level, tt.format, tt.args...)
		if !strings.Contains(buf.String(), tt.contain) {
			t.Errorf("Fprintf(%v, %q, %v) = %q, want to contain %q",
				tt.level, tt.format, tt.args, buf.String(), tt.contain)
		}
	}
}

func TestSpinner(t *testing.T) {
	oldColor := ColorEnabled
	ColorEnabled = false
	defer func() { ColorEnabled = oldColor }()

	s := NewSpinner("testing...")
	s.Start()
	s.Message("still testing...")
	s.Stop()
	// Should not panic or hang
}

func TestTable(t *testing.T) {
	oldColor := ColorEnabled
	ColorEnabled = false
	defer func() { ColorEnabled = oldColor }()

	var buf bytes.Buffer
	tbl := NewTable("Name", "Version", "Status")
	tbl.Row("mypkg", "v1.0.0", "ok")
	tbl.Row("other", "v2.0.0", "outdated")
	tbl.Fprint(&buf)

	out := buf.String()
	if !strings.Contains(out, "Name") {
		t.Error("table should contain header 'Name'")
	}
	if !strings.Contains(out, "mypkg") {
		t.Error("table should contain row 'mypkg'")
	}
}

func TestColorHelpers(t *testing.T) {
	oldColor := ColorEnabled
	ColorEnabled = false
	defer func() { ColorEnabled = oldColor }()

	if Bold("test") != "test" {
		t.Error("Bold should return plain text when colors disabled")
	}
	if Green("ok") != "ok" {
		t.Error("Green should return plain text when colors disabled")
	}
	if Red("err") != "err" {
		t.Error("Red should return plain text when colors disabled")
	}
}
