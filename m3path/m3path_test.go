package m3path

import (
	"github.com/freddy33/qsm-go/m3point"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPathContextFilling(t *testing.T) {
	for _, ctxType := range m3point.GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			pathCtx := PathContext{}
			pathCtx.ctx = m3point.GetTrioIndexContext(ctxType, pIdx)
			fillPathContext(t, &pathCtx, 5)
			break
		}
	}

}

func fillPathContext(t *testing.T, pathCtx *PathContext, until int) {
	Log.SetTrace()
	Log.SetAssert(true)
	m3point.Log.SetTrace()
	m3point.Log.SetAssert(true)

	trCtx := pathCtx.ctx
	trIdx := trCtx.GetBaseTrioIndex(0, 0)
	assert.NotEqual(t, m3point.NilTrioIndex, trIdx)

	td := m3point.GetTrioDetails(trIdx)
	assert.NotNil(t, td)
	assert.Equal(t, trIdx, td.GetId())

	Log.Debug(trCtx.String(), td.String())

	pathCtx.rootTrioId = trIdx
	for i, c := range td.GetConnections() {
		pathCtx.rootPathLinks[i] = makeRootPathLink(c.GetId())
	}
	for _, pl := range pathCtx.rootPathLinks {
		p, nextTrio, nextPathEl := trCtx.GetNextTrio(m3point.Origin, td, pl.connId)
		assert.NotNil(t, nextTrio, "Failed getting next trio for %s %s %s %v", trCtx.String(), td.String(), pl.String(), p)
		pl.dst = makePathNode(pl, nextTrio.GetId())
		for j := 0; j < 2; j++ {
			assert.True(t, nextPathEl[j].IsValid(), "Got invalid next path element for %s %s %s %v %v", trCtx.String(), td.String(), pl.String(), p, *nextPathEl[j])
			pl.dst.next[j] = makePathLink(pl.dst, nextPathEl[j].GetP2IConn().GetId())
		}
	}
	Log.Debug(pathCtx.dumpInfo())
}

func TestNilPathElement(t *testing.T) {
	nsp := EndPathElement(-3)
	assert.Equal(t, 0, nsp.GetLength())
	assert.Equal(t, int8(-3), nsp.GetForwardConnId(0))
	assert.Equal(t, nil, nsp.GetForwardElement(0))
	assert.Equal(t, nsp, nsp.Copy())
	assert.Equal(t, 0, TheEnd.GetLength())
	assert.Equal(t, int8(0), TheEnd.GetForwardConnId(0))
	assert.Equal(t, nil, TheEnd.GetForwardElement(0))
	assert.Equal(t, TheEnd, TheEnd.Copy())
}

func TestSimplePathElement(t *testing.T) {
	Log.SetDebug()
	sp1 := SimplePathElement{-2, nil,}
	assert.Equal(t, 1, sp1.NbForwardElements())
	assert.Equal(t, int8(-2), sp1.GetForwardConnId(0))
	assert.Equal(t, nil, sp1.GetForwardElement(0))
	assert.Equal(t, 1, sp1.GetLength())
	sp1Copy := sp1.Copy()
	//assert.NotEqual(t, &sp1, sp1Copy)
	assert.Equal(t, 1, sp1Copy.NbForwardElements())
	assert.Equal(t, int8(-2), sp1Copy.GetForwardConnId(0))
	assert.Equal(t, nil, sp1Copy.GetForwardElement(0))
	assert.Equal(t, 1, sp1Copy.GetLength())
	sp2 := SimplePathElement{-3, sp1Copy}
	assert.Equal(t, 2, sp2.GetLength())
	last := SimplePathElement{4, nil}
	sp2.SetLastNext(&last)
	assert.Equal(t, 1, last.GetLength())
	assert.Equal(t, 3, sp2.GetLength())
	assert.Equal(t, int8(-3), sp2.GetForwardConnId(0))
	assert.Equal(t, int8(4), last.GetForwardConnId(0))
	assert.Equal(t, nil, last.GetForwardElement(0))
	assert.Equal(t, &last, sp1Copy.GetForwardElement(0))
	assert.Equal(t, nil, sp1.GetForwardElement(0))
	sp1.next = &sp2
	assert.Equal(t, 4, sp1.GetLength())
	pathIds := []int8{-2, -3, -2, 4}
	var currentPath PathElement
	currentPath = &sp1
	for _, id := range pathIds {
		assert.Equal(t, id, currentPath.GetForwardConnId(0))
		currentPath = currentPath.GetForwardElement(0)
	}
}

func TestForkPathElement(t *testing.T) {
	Log.SetDebug()
	fp1 := ForkPathElement{make([]*SimplePathElement, 2)}
	fp1.simplePaths[0] = &SimplePathElement{-2, nil,}
	fp1.simplePaths[1] = &SimplePathElement{-3, nil,}
	assert.Equal(t, 2, fp1.NbForwardElements())
	assert.Equal(t, 1, fp1.GetLength())
	assert.Equal(t, int8(-2), fp1.GetForwardConnId(0))
	assert.Equal(t, int8(-3), fp1.GetForwardConnId(1))
	assert.Equal(t, nil, fp1.GetForwardElement(0))
	assert.Equal(t, nil, fp1.GetForwardElement(1))
	last := &SimplePathElement{4, nil}
	fp1.SetLastNext(last)
	assert.Equal(t, last, fp1.GetForwardElement(0))
	assert.Equal(t, last, fp1.GetForwardElement(1))
	assert.Equal(t, 2, fp1.GetLength())
}

func TestPathMergingErrors(t *testing.T) {
	assert.Equal(t, nil, MergePath(nil, nil))
	sp1 := &SimplePathElement{-2, nil,}
	sp2 := &SimplePathElement{+2, nil,}
	assert.Equal(t, nil, MergePath(sp1, nil))
	assert.Equal(t, nil, MergePath(nil, sp2))

	assert.Equal(t, 1, sp1.GetLength())
	assert.Equal(t, 1, sp2.GetLength())
	m := MergePath(sp1, sp2)
	assert.Equal(t, 1, m.GetLength())
	assert.Equal(t, 2, m.NbForwardElements())
	m1 := MergePath(sp1, sp1.Copy())
	assert.Equal(t, 1, m1.GetLength())
	assert.Equal(t, 1, m1.NbForwardElements())
	m2 := MergePath(sp2.Copy(), sp2)
	assert.Equal(t, 1, m2.GetLength())
	assert.Equal(t, 1, m2.NbForwardElements())
	sp1.SetLastNext(sp2)
	assert.Equal(t, nil, MergePath(sp1, sp2))
}

func TestPathMerging(t *testing.T) {
	Log.SetDebug()

	// Test Simple Merge
	path1Ids := []int8{-2, -3, -2, 4, 1, 1}
	path2Ids := []int8{-2, -3, +2, 4, 2, 2}
	var path1, path2 PathElement
	pLength := len(path1Ids)
	for i := pLength - 1; i >= 0; i-- {
		path1 = &SimplePathElement{path1Ids[i], path1}
		path2 = &SimplePathElement{path2Ids[i], path2}
	}
	assert.Equal(t, pLength, path1.GetLength())
	assert.Equal(t, pLength, path2.GetLength())
	merged := MergePath(path1, path2)
	assert.Equal(t, pLength, merged.GetLength())
	current := merged
	assert.Equal(t, pLength, merged.GetLength())
	assert.Equal(t, 1, current.NbForwardElements())
	assert.Equal(t, int8(-2), current.GetForwardConnId(0))
	current = current.GetForwardElement(0)
	pLength--
	assert.Equal(t, pLength, current.GetLength())
	assert.Equal(t, 1, current.NbForwardElements())
	assert.Equal(t, int8(-3), current.GetForwardConnId(0))
	current = current.GetForwardElement(0)
	pLength--
	assert.Equal(t, pLength, current.GetLength())
	assert.Equal(t, 2, current.NbForwardElements())
	assert.Equal(t, int8(-2), current.GetForwardConnId(0))
	assert.Equal(t, int8(2), current.GetForwardConnId(1))
	n1 := current.GetForwardElement(0)
	n2 := current.GetForwardElement(1)
	pLength--
	assert.Equal(t, pLength, n1.GetLength())
	assert.Equal(t, pLength, n2.GetLength())
	assert.Equal(t, int8(4), n1.GetForwardConnId(0))
	assert.Equal(t, int8(4), n2.GetForwardConnId(0))
	n1 = n1.GetForwardElement(0)
	n2 = n2.GetForwardElement(0)
	pLength--
	assert.Equal(t, pLength, n1.GetLength())
	assert.Equal(t, pLength, n2.GetLength())
	assert.Equal(t, int8(1), n1.GetForwardConnId(0))
	assert.Equal(t, int8(2), n2.GetForwardConnId(0))
	n1 = n1.GetForwardElement(0)
	n2 = n2.GetForwardElement(0)
	pLength--
	assert.Equal(t, pLength, n1.GetLength())
	assert.Equal(t, pLength, n2.GetLength())
	assert.Equal(t, int8(1), n1.GetForwardConnId(0))
	assert.Equal(t, int8(2), n2.GetForwardConnId(0))
	assert.Equal(t, nil, n1.GetForwardElement(0))
	assert.Equal(t, nil, n2.GetForwardElement(0))

	// Test Merge Actually Copy paths
	sp1 := &SimplePathElement{-2, nil,}
	merged.SetLastNext(sp1)
	pLength = len(path1Ids)
	assert.Equal(t, pLength, path1.GetLength())
	assert.Equal(t, pLength, path2.GetLength())
	assert.Equal(t, pLength+1, merged.GetLength())

	// Test more complex merge
	path3Ids := []int8{-2, -3, +2, 4, 3, -3}
	path4Ids := []int8{-2, -3, +2, 4, 5, -5}
	path5Ids := []int8{-2, -3, +2, 4, 6, -6}
	pLength = len(path3Ids)
	var path3, path4, path5 PathElement
	for i := pLength - 1; i >= 0; i-- {
		path3 = &SimplePathElement{path3Ids[i], path3}
		path4 = &SimplePathElement{path4Ids[i], path4}
		path5 = &SimplePathElement{path5Ids[i], path5}
	}
	assert.Equal(t, pLength, path3.GetLength())
	assert.Equal(t, pLength, path4.GetLength())
	assert.Equal(t, pLength, path5.GetLength())
	m34 := MergePath(path3, path4)
	m35 := MergePath(path3, path5)
	m45 := MergePath(path4, path5)
	assert.Equal(t, pLength, m34.GetLength())
	assert.Equal(t, pLength, m35.GetLength())
	assert.Equal(t, pLength, m45.GetLength())
	m345 := MergePath(m34, path5)
	m453 := MergePath(m45, path3)
	m3435 := MergePath(m34, m35)
	currentList := []PathElement{path3, path4, path5, m34, m35, m45, m345, m453, m3435}
	for j, c := range currentList {
		assert.NotEqual(t, nil, c, "c nil at %d", j)
		assert.Equal(t, pLength, c.GetLength(), "wrong length at %d", j)
	}
	// First 4 all identical
	for i := 0; i < 4; i++ {
		for j, c := range currentList {
			assert.NotEqual(t, nil, c, "c nil at %d %d", i, j)
			assert.Equal(t, 1, c.NbForwardElements(), "wrong nb elements at %d %d", i, j)
			assert.Equal(t, path3Ids[i], c.GetForwardConnId(0), "wrong conn ID at %d %d", i, j)
			assert.Equal(t, pLength-i, c.GetLength(), "wrong length at %d %d", i, j)
			currentList[j] = c.GetForwardElement(0)
		}
	}
	for j, c := range currentList {
		assert.NotEqual(t, nil, c, "c nil at %d", j)
		assert.Equal(t, pLength-4, c.GetLength(), "wrong length at %d", j)
	}

	assertSameConnList(t, currentList[0], map[int8]bool{3: true})
	assertSameConnList(t, currentList[1], map[int8]bool{5: true})
	assertSameConnList(t, currentList[2], map[int8]bool{6: true})
	assertSameConnList(t, currentList[3], map[int8]bool{3: true, 5: true})
	assertSameConnList(t, currentList[4], map[int8]bool{3: true, 6: true})
	assertSameConnList(t, currentList[5], map[int8]bool{5: true, 6: true})
	assertSameConnList(t, currentList[6], map[int8]bool{3: true, 5: true, 6: true})
	assertSameConnList(t, currentList[7], map[int8]bool{3: true, 5: true, 6: true})
	assertSameConnList(t, currentList[8], map[int8]bool{3: true, 5: true, 6: true})

	for j, c := range currentList {
		nbNext := c.NbForwardElements()
		for i := 0; i < nbNext; i++ {
			currentConnId := c.GetForwardConnId(i)
			next := c.GetForwardElement(i)
			assert.NotEqual(t, nil, next, "c nil at %d %d", i, j)
			assert.Equal(t, pLength-5, next.GetLength(), "wrong length at %d %d", i, j)
			assert.Equal(t, 1, next.NbForwardElements(), "wrong next nb elements length at %d %d", i, j)
			assert.Equal(t, next.GetForwardConnId(0), -currentConnId, "wrong connId at %d %d", i, j)
		}
	}
}

func assertSameConnList(t *testing.T, path PathElement, ids map[int8]bool) {
	assert.Equal(t, len(ids), path.NbForwardElements())
	for i := 0; i < len(ids); i++ {
		connId := path.GetForwardConnId(i)
		_, ok := ids[connId]
		assert.True(t, ok, "did not find connId %d from path %v in %v", connId, path, ids)
	}
}
