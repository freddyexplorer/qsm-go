package m3db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/freddy33/qsm-go/m3util"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

var Log = m3util.NewLogger("m3db", m3util.INFO)

type QsmEnvID int

const (
	NoEnv        QsmEnvID = iota // 0
	MainEnv                      // 1
	RunEnv                       // 2
	PerfTestEnv                  // 3
	ShellEnv                     // 4
	PointTestEnv                 // 5
	PathTestEnv                  // 6
	SpaceTestEnv                 // 7
	GlTestEnv                    // 8
	DbTempEnv                    // 9
	PointTempEnv                 // 10
	PathTempEnv                  // 11
	PointLoadEnv                 // 12
)

const (
	MaxNumberOfEnvironments = 15
	QsmEnvNumberKey         = "QSM_ENV_NUMBER"
)

type DbConnDetails struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbName"`
}

type QsmError string

func MakeQsmErrorf(format string, args ...interface{}) QsmError {
	return QsmError(fmt.Sprintf(format, args...))
}

func (qsmError QsmError) Error() string {
	return string(qsmError)
}

const (
	PointIdx     = 0
	PathIdx      = 1
	SpaceIdx     = 2
	GlIdx        = 3
	MaxDataEntry = 4
)

type QsmEnvironment struct {
	id               QsmEnvID
	dbDetails        DbConnDetails
	db               *sql.DB
	createTableMutex sync.Mutex
	tableExecs       map[string]*TableExec
	data             [MaxDataEntry]interface{}
}

var createEnvMutex sync.Mutex
var environments map[QsmEnvID]*QsmEnvironment

var TestMode bool

func init() {
	environments = make(map[QsmEnvID]*QsmEnvironment)
}

func SetToTestMode() {
	TestMode = true
}

func GetDefaultEnvId() QsmEnvID {
	if TestMode {
		Log.Fatalf("Cannot use not set pointEnv in test mode!")
	}
	envId := MainEnv
	envIdFromOs := os.Getenv(QsmEnvNumberKey)
	if envIdFromOs != "" {
		id, err := strconv.ParseInt(envIdFromOs, 10, 16)
		if err != nil {
			Log.Fatalf("The %s environment variable is not a DB number but %s", QsmEnvNumberKey, envIdFromOs)
		}
		envId = QsmEnvID(id)
	}
	Log.Infof("Using default environment %d", envId)
	return envId
}

func GetDefaultEnvironment() *QsmEnvironment {
	if TestMode {
		Log.Fatalf("Cannot use default environment in test mode!")
	}
	return GetEnvironment(GetDefaultEnvId())
}

func GetEnvironment(envId QsmEnvID) *QsmEnvironment {
	env, ok := environments[envId]
	if !ok {
		createEnvMutex.Lock()
		defer createEnvMutex.Unlock()
		env, ok = environments[envId]
		if !ok {
			env = createNewEnv(envId)
			environments[envId] = env
		}
	}
	return env
}

func RemoveEnvFromMap(envId QsmEnvID) {
	createEnvMutex.Lock()
	defer createEnvMutex.Unlock()
	delete(environments, envId)
}

func (env *QsmEnvironment) GetId() QsmEnvID {
	return env.id
}

func (env *QsmEnvironment) GetConnection() *sql.DB {
	return env.db
}

func (env *QsmEnvironment) GetDbConf() DbConnDetails {
	return env.dbDetails
}

func createNewEnv(envId QsmEnvID) *QsmEnvironment {
	env := QsmEnvironment{}
	env.id = envId
	env.tableExecs = make(map[string]*TableExec)

	env.checkOsEnv()
	env.fillDbConf()
	env.openDb()

	env.Ping()

	return &env
}

func SetEnvQuietly(key, value string) {
	m3util.ExitOnError(os.Setenv(key, value))
}

func (env *QsmEnvironment) GetEnvNumber() string {
	return strconv.Itoa(int(env.id))
}

func (env *QsmEnvironment) checkOsEnv() {
	envNumber := env.GetEnvNumber()
	origQsmId := os.Getenv(QsmEnvNumberKey)

	if envNumber != origQsmId {
		// Reset the env var to what it was on exit of this method
		defer SetEnvQuietly(QsmEnvNumberKey, origQsmId)
		// set the env var correctly
		m3util.ExitOnError(os.Setenv(QsmEnvNumberKey, envNumber))
	}

	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "db", "check")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Fatalf("failed to check environment %d at OS level due to %v with output: ***\n%s\n***", env.id, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("check environment %d at OS output: ***\n%s\n***", env.id, string(out))
		}
	}
}

func (env *QsmEnvironment) fillDbConf() {
	connJsonFile := fmt.Sprintf("%s/dbconn%d.json", m3util.GetConfDir(), env.id)
	confData, err := ioutil.ReadFile(connJsonFile)
	if err != nil {
		log.Fatalf("failed opening DB conf file %s due to %v", connJsonFile, err)
	}
	err = json.Unmarshal([]byte(confData), &env.dbDetails)
	if err != nil {
		log.Fatalf("failed parsing DB conf file %s due to %v", connJsonFile, err)
	}
	if Log.IsDebug() {
		Log.Debugf("DB conf for environment %d is user=%s dbName=%s", env.id, env.dbDetails.User, env.dbDetails.DbName)
	}
}

func (env *QsmEnvironment) openDb() {
	connDetails := env.GetDbConf()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		connDetails.Host, connDetails.Port, connDetails.User, connDetails.Password, connDetails.DbName)
	if Log.IsDebug() {
		Log.Debugf("Opening DB for environment %d is user=%s dbName=%s", env.id, env.dbDetails.User, env.dbDetails.DbName)
	}
	var err error
	env.db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		Log.Fatalf("fail to open DB for environment %d with user=%s and dbName=%s due to %v", env.id, env.dbDetails.User, env.dbDetails.DbName, err)
	}
	if Log.IsDebug() {
		Log.Debugf("DB opened for environment %d is user=%s dbName=%s", env.id, env.dbDetails.User, env.dbDetails.DbName)
	}
}

func (env *QsmEnvironment) _internalClose() error {
	envId := env.id
	defer RemoveEnvFromMap(envId)
	// clean data
	for i := 0; i < len(env.data); i++ {
		env.data[i] = nil
	}
	// clean table exec
	for tn, te := range env.tableExecs {
		err := te.Close()
		if err != nil {
			Log.Warnf("Closing table exec of envId=%d table=%s generated '%s'", env.id, tn, err.Error())
		}
		delete(env.tableExecs, tn)
	}
	db := env.db
	env.db = nil
	if db != nil {
		return db.Close()
	}
	return nil
}

func CloseAll() {
	toClose := make([]*QsmEnvironment, 0, len(environments))
	for _, env := range environments {
		if env != nil {
			toClose = append(toClose, env)
		}
	}
	for _, env := range toClose {
		CloseEnv(env)
	}
}

func CloseEnv(env *QsmEnvironment) {
	if env == nil {
		Log.Warn("Closing nil environment")
		return
	}
	m3util.ExitOnError(env._internalClose())
}

func (env *QsmEnvironment) Destroy() {
	envId := env.id
	err := env._internalClose()
	if err != nil {
		Log.Error(err)
	}

	envNumber := env.GetEnvNumber()
	origQsmId := os.Getenv(QsmEnvNumberKey)

	if envNumber != origQsmId {
		// Reset the env var to what it was on exit of this method
		defer SetEnvQuietly(QsmEnvNumberKey, origQsmId)
		// set the env var correctly
		m3util.ExitOnError(os.Setenv(QsmEnvNumberKey, envNumber))
	}

	rootDir := m3util.GetGitRootDir()
	cmd := exec.Command("bash", filepath.Join(rootDir, "qsm"), "db", "drop")
	out, err := cmd.CombinedOutput()
	if err != nil {
		Log.Errorf("failed to destroy environment %d at OS level due to %v with output: ***\n%s\n***", envId, err, string(out))
	} else {
		if Log.IsDebug() {
			Log.Debugf("destroy environment %d at OS level output: ***\n%s\n***", envId, string(out))
		}
	}
}

func (env *QsmEnvironment) Ping() bool {
	err := env.GetConnection().Ping()
	if err != nil {
		Log.Errorf("failed to ping %d on DB %s due to %v", env.id, env.dbDetails.DbName, err)
		return false
	}
	if Log.IsDebug() {
		Log.Debugf("ping for environment %d successful", env.id)
	}
	return true
}

func (env *QsmEnvironment) GetData(dataIdx int) interface{} {
	return env.data[dataIdx]
}

func (env *QsmEnvironment) SetData(dataIdx int, d interface{}) {
	env.data[dataIdx] = d
}
