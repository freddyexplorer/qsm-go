package m3db

import (
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDbConf(t *testing.T) {
	confDir := m3util.GetConfDir()
	assert.True(t, strings.HasSuffix(confDir, "conf"), "conf dir %s does end with conf", confDir)

	cmd := exec.Command("cp", filepath.Join(confDir, "db-test.json"), filepath.Join(confDir, "dbconn1234.json"))
	err := cmd.Run()
	m3util.ExitOnError(err)

	connDetails := readDbConf(ConfEnv)
	assert.Equal(t, "hostTest", connDetails.Host, "fails reading %v", connDetails)
	assert.Equal(t, 1234, connDetails.Port, "fails reading %v", connDetails)
	assert.Equal(t, "userTest", connDetails.User, "fails reading %v", connDetails)
	assert.Equal(t, "passwordTest", connDetails.Password, "fails reading %v", connDetails)
	assert.Equal(t, "dbNameTest", connDetails.DbName, "fails reading %v", connDetails)
}

func TestDbConnection(t *testing.T) {
	defer DropEnv(TempEnv)
	CheckOrCreateEnv(TempEnv)
	db := GetConnection(TempEnv)
	defer CloseDb(db)
	err := db.Ping()
	assert.True(t, err == nil, "Got ping error %v", err)
}
