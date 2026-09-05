package bim

import "testing"

func TestValidateFormat(t *testing.T) {
	for _, v := range []string{"ifc", "gltf", "glb", "obj", "dxf", "native"} {
		if err := ValidateFormat(v); err != nil {
			t.Fatalf("%s: %v", v, err)
		}
	}
	if err := ValidateFormat("pdf"); err == nil {
		t.Fatal("expected invalid format")
	}
}
func TestValidateStatus(t *testing.T) {
	if err := ValidateStatus("approved"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateStatus("deleted"); err == nil {
		t.Fatal("expected invalid status")
	}
}
func TestValidateProgress(t *testing.T) {
	if err := ValidateProgress(50, 40); err != nil {
		t.Fatal(err)
	}
	if err := ValidateProgress(-1, 20); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateProgress(20, 101); err == nil {
		t.Fatal("expected error")
	}
}
func TestValidateVersionAndIO(t *testing.T) {
	if err := ValidateVersion(1); err != nil {
		t.Fatal(err)
	}
	if err := ValidateVersion(0); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateImportExport("import", "ifc"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateImportExport("delete", "ifc"); err == nil {
		t.Fatal("expected error")
	}
}
