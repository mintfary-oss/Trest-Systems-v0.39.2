package importer

import (
	"context"
	"database/sql"
	"github.com/mintfary-oss/trest-sistems/internal/bim/ifc"
	"strings"
	"testing"
)

func TestImportRequiresDB(t *testing.T) {
	es, _ := ifc.ParseSTEP(strings.NewReader("#1=IFCWALL('g',$,'Wall',$,$,$,$,$,$,.NOTDEFINED.);"))
	if _, err := ImportIFC(context.Background(), nil, "v", es); err == nil {
		t.Fatal("expected nil db error")
	}
	_ = sql.ErrNoRows
}
