package m3path

import (
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var allTestContextsMutex sync.Mutex

func getAllTestContexts(env *m3db.QsmEnvironment) map[m3point.GrowthType][]PathContext {
	pathData := GetPathPackData(env)
	if pathData.allCenterContextsLoaded {
		return pathData.allCenterContexts
	}

	allTestContextsMutex.Lock()
	defer allTestContextsMutex.Unlock()

	if pathData.allCenterContextsLoaded {
		return pathData.allCenterContexts
	}

	m3point.InitializeDBEnv(env, false)
	pointData := m3point.GetPointPackData(env)

	idx := 0
	for _, growthCtx := range pointData.GetAllGrowthContexts() {
		ctxType := growthCtx.GetGrowthType()
		maxOffset := ctxType.GetMaxOffset()
		if len(pathData.allCenterContexts[ctxType]) == 0 {
			pathData.allCenterContexts[ctxType] = make([]PathContext, ctxType.GetNbIndexes()*maxOffset)
			idx = 0
		}
		for offset := 0; offset < maxOffset; offset++ {
			pathData.allCenterContexts[ctxType][idx] = MakePathContextDBFromGrowthContext(env, growthCtx, offset)
			idx++
		}
	}

	pathData.allCenterContextsLoaded = true
	return pathData.allCenterContexts
}

func TestFirstPathContextFilling(t *testing.T) {
	Log.SetInfo()
	Log.SetAssert(true)
	m3point.Log.SetInfo()
	m3point.Log.SetAssert(true)
	m3db.SetToTestMode()

	env := GetFullTestDb(m3db.PathTestEnv)
	allCtx := getAllTestContexts(env)
	for _, ctxType := range m3point.GetAllContextTypes() {
		for _, ctx := range allCtx[ctxType] {
			fillPathContext(t, ctx, 8*3)
			Log.Infof("Run for %s got %d points %d last open end path", ctx.String(), ctx.CountAllPathNodes(), ctx.GetNumberOfOpenNodes())
			Log.Debug(ctx.dumpInfo())
			break
		}
	}
}

func fillPathContext(t *testing.T, pathCtx PathContext, until int) {
	growthCtx := pathCtx.GetGrowthCtx()
	trIdx := growthCtx.GetBaseTrioIndex(0, pathCtx.GetGrowthOffset())
	assert.NotEqual(t, m3point.NilTrioIndex, trIdx)

	ppd := m3point.GetPointPackData(pathCtx.GetGrowthCtx().GetEnv())

	td := ppd.GetTrioDetails(trIdx)
	assert.NotNil(t, td)
	assert.Equal(t, trIdx, td.GetId())

	Log.Debug(growthCtx.String(), td.String())

	pathCtx.InitRootNode(m3point.Origin)
	pathCtx.MoveToNextNodes()

	//pathNodeMap := pathCtx.GetPathNodeMap()
	assert.Equal(t, 1+3, pathCtx.CountAllPathNodes(), "not all points of %s are here", pathCtx.String())
	assert.Equal(t, 3, pathCtx.GetNumberOfOpenNodes(), "not all ends of %s here", pathCtx.String())
	//spnm, ok := pathNodeMap.(*SimplePathNodeMap)
	//assert.True(t, ok, "should be a simple path node map for %v", pathNodeMap)
	countMains := 0
	countNonMains := 0
	openEndNodes := pathCtx.GetAllOpenPathNodes()
	for _, pn := range openEndNodes {
		assert.NotEqual(t, m3point.NilTrioIndex, pn.GetTrioIndex(), "%v should have trio already", pn)
		if pn.P().IsMainPoint() {
			countMains++
		} else {
			countNonMains++
		}
		assert.Equal(t, 1, pn.D(), "open end path %v should have distance of three", pn)
		assert.True(t, pn.IsLatest(), "open end path %v should be active", pn)
	}
	assert.Equal(t, 0, countMains, "not all main ends here %v", openEndNodes)
	assert.Equal(t, 3, countNonMains, "not all non main ends here %v", openEndNodes)

	if until == 2 {
		Log.Debug("*************** First round *************\n", pathCtx.dumpInfo())
		pathCtx.MoveToNextNodes()
		assertPathContextState(t, pathCtx.GetAllOpenPathNodes())
		Log.Debug("*************** Second round *************\n", pathCtx.dumpInfo())
	} else {
		for d := 1; d < until; d++ {
			pathCtx.MoveToNextNodes()
			assertPathContextState(t, pathCtx.GetAllOpenPathNodes())
		}
	}
}

func assertPathContextState(t *testing.T, openEndNodes []PathNode) {
	//inOpenEnd := make(map[m3point.Point]bool)
	for _, pn := range openEndNodes {
		assert.True(t, pn.(*PathNodeDb).id > 0, "%v should have an id already", pn)
		assert.NotEqual(t, m3point.NilTrioIndex, pn.GetTrioIndex(), "%v should have trio already", pn)
		//assert.Equal(t, pn.calcDist(), pn.D(), "open end path %v should have d and calcDist equal", pn)
		// TODO: Find a way to test that open end node are mostly active
		//assert.True(t, oep.pn.IsLatest(), "open end path %v should be active", oep.pn)
		//inOpenEnd[pn.P()] = true
	}
	//for p, n := range *spnm {
	//	if !inOpenEnd[p] {
	//		assert.False(t, n.IsLatest(), "non open end path %v should be active", n)
	//	}
	//}
}
