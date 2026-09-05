package api

import "testing"

func TestValidateIFCFilename(t *testing.T) {
	valid := []string{"model.ifc", "MODEL.IFC", "building-01.ifc"}
	for _, name := range valid {
		got, err := validateIFCFilename(name)
		if err != nil || got != name {
			t.Fatalf("valid filename %q rejected: got=%q err=%v", name, got, err)
		}
	}
	invalid := []string{"../model.ifc", `..\\model.ifc`, `/tmp/model.ifc`, "model.txt", "", ".", "model.ifc\x00"}
	for _, name := range invalid {
		if _, err := validateIFCFilename(name); err == nil {
			t.Fatalf("invalid filename %q accepted", name)
		}
	}
}
