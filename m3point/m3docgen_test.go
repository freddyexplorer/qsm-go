package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"testing"
)

func TestWriteAllTables(t *testing.T) {
	GenerateTextFilesEnv(GetFullTestDb(m3db.PointTestEnv))
}