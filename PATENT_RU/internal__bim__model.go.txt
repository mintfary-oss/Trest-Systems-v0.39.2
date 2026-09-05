package bim

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidFormat   = errors.New("invalid BIM format")
	ErrInvalidStatus   = errors.New("invalid BIM status")
	ErrInvalidProgress = errors.New("progress must be between 0 and 100")
	ErrInvalidVersion  = errors.New("version must be greater than zero")
)

var formats = map[string]bool{"native": true, "ifc": true, "gltf": true, "glb": true, "obj": true, "dxf": true}
var statuses = map[string]bool{"draft": true, "review": true, "approved": true, "archived": true}

type Model struct{ ID, ProjectID, ProjectVersionID, Name, Format, StorageURL, SchemaVersion, Status string }
type ModelVersion struct {
	ID, BIMModelID, SourceFormat, SourceURI, GeometryURI, Checksum string
	Version                                                        int
}
type Element struct{ ID, BIMModelVersionID, ExternalID, ElementType, Name string }
type ProgressSnapshot struct {
	ProjectID, BIMModelVersionID, Source string
	PlannedPercent, ActualPercent        float64
}

type ImportExport struct{ Operation, Format, Status string }

func ValidateFormat(v string) error {
	if !formats[strings.ToLower(strings.TrimSpace(v))] {
		return fmt.Errorf("%w: %s", ErrInvalidFormat, v)
	}
	return nil
}
func ValidateStatus(v string) error {
	if !statuses[strings.ToLower(strings.TrimSpace(v))] {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, v)
	}
	return nil
}
func ValidateVersion(v int) error {
	if v <= 0 {
		return ErrInvalidVersion
	}
	return nil
}
func ValidateProgress(planned, actual float64) error {
	if planned < 0 || planned > 100 || actual < 0 || actual > 100 {
		return ErrInvalidProgress
	}
	return nil
}
func ValidateImportExport(operation, format string) error {
	if operation != "import" && operation != "export" {
		return fmt.Errorf("invalid BIM operation: %s", operation)
	}
	return ValidateFormat(format)
}
