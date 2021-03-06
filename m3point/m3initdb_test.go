package m3point

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const(
	ExpectedNbConns = 50
	ExpectedNbTrios = 200
	ExpectedNbGrowthContexts = 52
	ExpectedNbCubes = 5192
	ExpectedNbPathBuilders = ExpectedNbCubes
)

func TestLoadOrCalculate(t *testing.T) {
	m3db.Log.SetInfo()
	Log.SetInfo()
	m3db.SetToTestMode()

	env := GetFullTestDb(m3db.PointLoadEnv)
	ppd := GetPointPackData(env)

	start := time.Now()
	ppd.resetFlags()
	ppd.allConnections, ppd.allConnectionsByVector = ppd.calculateConnectionDetails()
	ppd.connectionsLoaded = true
	ppd.allTrioDetails = ppd.calculateAllTrioDetails()
	ppd.trioDetailsLoaded = true
	ppd.allGrowthContexts = ppd.calculateAllGrowthContexts()
	ppd.growthContextsLoaded = true
	ppd.cubeIdsPerKey = ppd.calculateAllContextCubes()
	ppd.cubesLoaded = true
	ppd.pathBuilders = ppd.calculateAllPathBuilders()
	ppd.pathBuildersLoaded = true
	calcTime := time.Now().Sub(start)
	Log.Infof("Took %v to calculate", calcTime)

	assert.Equal(t, ExpectedNbConns, len(ppd.allConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.allConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.allTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.allGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.pathBuilders)-1)

	start = time.Now()
	// force reload
	InitializeDBEnv(env, true)
	loadTime := time.Now().Sub(start)
	Log.Infof("Took %v to load", loadTime)

	Log.Infof("Diff calc-load = %v", calcTime - loadTime)

	// Don't forget to get ppd different after init
	ppd = GetPointPackData(env)
	assert.Equal(t, ExpectedNbConns, len(ppd.allConnections))
	assert.Equal(t, ExpectedNbConns, len(ppd.allConnectionsByVector))
	assert.Equal(t, ExpectedNbTrios, len(ppd.allTrioDetails))
	assert.Equal(t, ExpectedNbGrowthContexts, len(ppd.allGrowthContexts))
	assert.Equal(t, ExpectedNbCubes, len(ppd.cubeIdsPerKey))
	assert.Equal(t, ExpectedNbPathBuilders, len(ppd.pathBuilders)-1)
}

func TestSaveAll(t *testing.T) {
	m3db.Log.SetDebug()
	Log.SetDebug()
	m3db.SetToTestMode()

	tempEnv := GetCleanTempDb(m3db.PointTempEnv)
	ppd := GetPointPackData(tempEnv)

	// ************ Connection Details

	n, err := ppd.saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Should be able to run twice
	n, err = ppd.saveAllConnectionDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbConns, n)

	// Test we can load
	loaded, _ := loadConnectionDetails(tempEnv)
	assert.Equal(t, ExpectedNbConns, len(loaded))

	// Init
	ppd.initConnections()

	// ************ Trio Details

	n, err = ppd.saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Should be able to run twice
	n, err = ppd.saveAllTrioDetails()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbTrios, n)

	// Test we can load
	loaded2 := ppd.loadTrioDetails()
	assert.Equal(t, ExpectedNbTrios, len(loaded2))

	// Init
	ppd.initTrioDetails()

	// ************ Growth Contexts

	n, err = ppd.saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Should be able to run twice
	n, err = ppd.saveAllGrowthContexts()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbGrowthContexts, n)

	// Test we can load
	loaded3 := ppd.loadGrowthContexts()
	assert.Equal(t, ExpectedNbGrowthContexts, len(loaded3))

	// Init
	ppd.initGrowthContexts()

	// ************ Context Cubes

	n, err = ppd.saveAllContextCubes()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Should be able to run twice
	n, err = ppd.saveAllContextCubes()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbCubes, n)

	// Test we can load
	loaded4 := ppd.loadContextCubes()
	assert.Equal(t, ExpectedNbCubes, len(loaded4))

	// Init
	ppd.initContextCubes()

	// ************ Path Builders

	n, err = ppd.saveAllPathBuilders()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Should be able to run twice
	n, err = ppd.saveAllPathBuilders()
	assert.Nil(t, err)
	assert.Equal(t, ExpectedNbPathBuilders, n)

	// Test we can load
	loaded5 := ppd.loadPathBuilders()
	assert.Equal(t, ExpectedNbPathBuilders, len(loaded5)-1)

	// Init from Good DB
	ppd.initPathBuilders()
}
