package m3point

import (
	"github.com/freddy33/qsm-go/m3util"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
)

const (
	SPLIT          = 4
	BENCH_NB_ROUND = 100
	TEST_NB_ROUND  = 25
)

func BenchmarkCtx1(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 1)
}

func BenchmarkCtx2(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 2)
}

func BenchmarkCtx3(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 3)
}

func BenchmarkCtx4(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 4)
}

func BenchmarkCtx8(b *testing.B) {
	Log.Level = m3util.WARN
	runForCtxType(b.N, BENCH_NB_ROUND, 8)
}

func TestCtx2(t *testing.T) {
	Log.Level = m3util.INFO
	runForCtxType(1, TEST_NB_ROUND, 2)
}

func TestCtxPerType(t *testing.T) {
	Log.Level = m3util.INFO
	for _, pType := range GetAllContextTypes() {
		runForCtxType(1, TEST_NB_ROUND, pType)
	}
}

func runForCtxType(N, nbRound int, pType ContextType) {
	allCtx := getAllTestContexts()
	for r := 0; r < N; r++ {
		maxUsed := 0
		maxLatest := 0
		for _, ctx := range allCtx[pType] {
			nU, nL := runNextPoints(&ctx, nbRound)
			if nU > maxUsed {
				maxUsed = nU
			}
			if nL > maxLatest {
				maxLatest = nL
			}
		}
		Log.Infof("Max size for all context of type %d: %d, %d with %d runs", pType, maxUsed, maxLatest, nbRound)
	}
}

func BenchmarkAllGrowth(b *testing.B) {
	Log.Level = m3util.WARN
	nbRound := 50
	allCtx := getAllTestContexts()
	for r := 0; r < b.N; r++ {
		maxUsed := 0
		maxLatest := 0
		for _, pType := range GetAllContextTypes() {
			for _, ctx := range allCtx[pType] {
				nU, nL := runNextPoints(&ctx, nbRound)
				if nU > maxUsed {
					maxUsed = nU
				}
				if nL > maxLatest {
					maxLatest = nL
				}
			}
		}
		Log.Infof("Max size for all context %d, %d with %d runs", maxUsed, maxLatest, nbRound)
	}
}

func runNextPoints(ctx *GrowthContext, nbRound int) (int, int) {
	usedPoints := make(map[Point]bool, 10*nbRound*nbRound)
	totalUsedPoints := 1
	latestPoints := make([]Point, 1)
	latestPoints[0] = Origin
	usedPoints[Origin] = true
	for d := 0; d < nbRound; d++ {
		nbLatestPoints := len(latestPoints)
		// Send all orig new points
		origNewPoints := make(chan Point, 4*SPLIT)
		wg := sync.WaitGroup{}
		if nbLatestPoints < 4*SPLIT {
			// too small for split send all
			wg.Add(1)
			go nextPointsSplit(&latestPoints, 0, nbLatestPoints, ctx, origNewPoints, &wg)
		} else {
			sizePerSplit := int(nbLatestPoints / SPLIT)
			for currentPos := 0; currentPos < nbLatestPoints; currentPos += sizePerSplit {
				wg.Add(1)
				go nextPointsSplit(&latestPoints, currentPos, sizePerSplit, ctx, origNewPoints, &wg)
			}
		}
		go func(step int) {
			wg.Wait()
			close(origNewPoints)
		}(d)

		finalPoints := make([]Point, 0, int(1.7*float32(nbLatestPoints)))
		for p := range origNewPoints {
			_, ok := usedPoints[p]
			if !ok {
				finalPoints = append(finalPoints, p)
				usedPoints[p] = true
			}
		}

		totalUsedPoints += len(finalPoints)
		latestPoints = finalPoints
	}
	return totalUsedPoints, len(latestPoints)
}

func runNextPointsAsync(ctx *GrowthContext, nbRound int) (int, int) {
	//usedPoints := make(map[Point]bool, 10*nbRound*nbRound)
	usedPoints := new(sync.Map)
	totalUsedPoints := 1
	latestPoints := make([]Point, 1)
	latestPoints[0] = Origin
	usedPoints.Store(Origin, true)
	o := make(chan Point, 100)
	for d := 0; d < nbRound; d++ {
		finalPoints := make([]Point, 0, int(1.2*float32(len(latestPoints))))
		for _, p := range latestPoints {
			go asyncNextPoints(p, ctx, o, nil)
		}
		// I'll always get 3 tines the amount of latest points
		newPoints := 3 * len(latestPoints)
		for i := 0; i < newPoints; i++ {
			p, ok := <-o
			if !ok {
				break
			} else {
				_, ok := usedPoints.LoadOrStore(p, true)
				if !ok {
					finalPoints = append(finalPoints, p)
				}
			}
		}
		latestPoints = finalPoints
		totalUsedPoints += len(latestPoints)
	}
	return totalUsedPoints, len(latestPoints)
}

func nextPointsSplit(lps *[]Point, currentPos, nb int, ctx *GrowthContext, o chan Point, wg *sync.WaitGroup) {
	c := 0
	for i := currentPos; i < len(*lps); i++ {
		p := (*lps)[i]
		for _, np := range p.GetNextPoints(ctx) {
			o <- np
		}
		c++
		if c == nb {
			break
		}
	}
	wg.Done()
}

func asyncNextPoints(p Point, ctx *GrowthContext, o chan Point, wg *sync.WaitGroup) {
	for _, np := range p.GetNextPoints(ctx) {
		o <- np
	}
	wg.Done()
}

var allTestContexts map[ContextType][]GrowthContext

func getAllTestContexts() map[ContextType][]GrowthContext {
	if allTestContexts != nil {
		return allTestContexts
	}
	res := make(map[ContextType][]GrowthContext)

	for _, ctxType := range GetAllContextTypes() {
		nbIndexes := ctxType.GetNbIndexes()
		maxOffset := maxOffsetPerType[ctxType]
		res[ctxType] = make([]GrowthContext, nbIndexes*maxOffset)
		idx := 0
		for pIdx := 0; pIdx < nbIndexes; pIdx++ {
			rootCtx := GetTrioIndexContext(ctxType, pIdx)
			for offset := 0; offset < maxOffset; offset++ {
				res[ctxType][idx] = *CreateFromRoot(rootCtx, Origin, offset)
				idx++
			}
		}
	}

	allTestContexts = res
	return res
}

func TestDivByThree(t *testing.T) {
	runDivByThree(t)
}

func runDivByThree(t assert.TestingT) {
	Log.Level = m3util.DEBUG
	someCenter1 := Point{3, -6, 9}
	ctx := CreateGrowthContext(someCenter1, 1, 1, 0)
	assert.Equal(t, someCenter1, ctx.center)
	assert.Equal(t, ContextType(1), ctx.ctxType)
	assert.Equal(t, 1, ctx.ctxIndex)
	assert.Equal(t, 0, ctx.offset)

	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{0, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{6, -6, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -3, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -9, 9}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -6, 12}))
	assert.Equal(t, uint64(1), ctx.GetDivByThree(Point{3, -6, 6}))

	assert.Equal(t, uint64(6), ctx.GetDivByThree(Point{0, 0, 0}))

	// Verify trio index unaffected
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 1, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}

}

func TestGrowthContext1(t *testing.T) {
	Log.Level = m3util.DEBUG
	ctx := CreateGrowthContext(Origin, 1, 3, 0)
	assert.Equal(t, ContextType(1), ctx.ctxType)
	assert.Equal(t, 3, ctx.ctxIndex)
	assert.Equal(t, 0, ctx.offset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 3, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
	ctx.ctxIndex = 4
	ctx.offset = 2
	assert.Equal(t, ContextType(1), ctx.ctxType)
	assert.Equal(t, 4, ctx.ctxIndex)
	assert.Equal(t, 2, ctx.offset)
	for d := uint64(0); d < 30; d++ {
		assert.Equal(t, 4, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx, d)
	}
}

func TestGrowthContext3(t *testing.T) {
	Log.Level = m3util.DEBUG

	for idx := 0; idx < 4; idx++ {
		ctx := CreateGrowthContext(Origin, 3, idx, 0)
		assert.Equal(t, ContextType(3), ctx.ctxType)
		assert.Equal(t, idx, ctx.ctxIndex)
		assert.Equal(t, 0, ctx.offset)
		for d := uint64(0); d < 9; d++ {
			if d%2 == 0 {
				assert.Equal(t, idx, ctx.GetTrioIndex(d), "failed trio index for ctx %v step %d", ctx, d)
			} else {
				expected := 4 + (int(d/2) % 3)
				if expected >= idx+4 {
					expected++
				}
				assert.Equal(t, expected, ctx.GetTrioIndex(d), "failed trio index for ctx %v step %d", ctx, d)
			}
		}
	}
}

func TestGrowthContextsExpectType3(t *testing.T) {
	runGrowthContextsExpectType3(t)
}

func runGrowthContextsExpectType3(t assert.TestingT) {
	Log.Level = m3util.DEBUG

	growthContexts := getAllTestContexts()
	for _, ctx := range growthContexts[1] {
		assert.Equal(t, ContextType(1), ctx.ctxType)
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, ctx.ctxIndex, ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d", ctx.String(), d)
		}
	}

	for _, ctx := range growthContexts[2] {
		assert.Equal(t, ContextType(2), ctx.ctxType)
		oneTwo := ValidNextTrio[ctx.ctxIndex]
		twoIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneTwo[twoIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d twoIdx=%d in %v", ctx.String(), d, twoIdx, oneTwo)
			twoIdx++
			if twoIdx == 2 {
				twoIdx = 0
			}
		}
	}

	for _, ctx := range growthContexts[4] {
		assert.Equal(t, ContextType(4), ctx.ctxType)
		oneToFour := AllMod4Permutations[ctx.ctxIndex]
		fourIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToFour[fourIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d fourIdx=%d in %v", ctx.String(), d, fourIdx, oneToFour)
			fourIdx++
			if fourIdx == 4 {
				fourIdx = 0
			}
		}
	}

	for _, ctx := range growthContexts[8] {
		assert.Equal(t, ContextType(8), ctx.ctxType)
		oneToEight := AllMod8Permutations[ctx.ctxIndex]
		eightIdx := ctx.offset
		for d := uint64(0); d < 30; d++ {
			assert.Equal(t, oneToEight[eightIdx], ctx.GetTrioIndex(d), "failed trio index for ctx %v and divByThree=%d eightIdx=%d in %v", ctx.String(), d, eightIdx, oneToEight)
			eightIdx++
			if eightIdx == 8 {
				eightIdx = 0
			}
		}
	}

}

func TestTrioListPerContext(t *testing.T) {
	Log.Level = m3util.INFO
	contexts := getAllTestContexts()
	for _, ctxs := range contexts {
		stableStep := -1
		indexList := make(map[int][]int)
		for _, ctx := range ctxs {
			s, l := runAllTrioList(t, &ctx)
			if stableStep == -1 {
				stableStep = s
			} else {
				assert.Equal(t, stableStep, s, "failed same stable step for %s", ctx.String())
			}
			curList, ok := indexList[ctx.ctxIndex]
			if !ok {
				indexList[ctx.ctxIndex] = l
			} else {
				assert.True(t, EqualIntSlice(curList, l), "failed same index list for %s %v != %v", ctx.String(), curList, l)
			}
		}
	}
}

// Return the ordered list of trio index used
func runAllTrioList(t *testing.T, ctx *GrowthContext) (stableStep int, indexList []int) {
	// The result list of trio index used
	var currentIndexList []int

	countUsedIdx := make(map[int]int)
	usedPoints := make(map[Point]int)
	latestPoints := make([]Point, 1)
	latestPoints[0] = ctx.center

	// If currentIndexList is stable for "verifyStable" iterations we should stop
	verifyStable := 8
	stableIndexList := 0
	stepStable := 0

	for d := uint64(0); d < 30; d++ {
		stepStable = int(d)
		newPoints := make([]Point, 0, 2*len(latestPoints))
		stepCountIdx := make(map[int]int)
		stepConflictCount := make(map[Point]int)

		for _, p := range latestPoints {
			nextPoints := p.GetNextPoints(ctx)
			tdIdx, link := findTrioIndex(p, nextPoints, ctx)
			if !p.IsMainPoint() {
				td := AllTrioDetails[tdIdx]
				foundLink := false
				for _, l := range td.links {
					if l == link {
						foundLink = true
						break
					}
				}
				assert.True(t, foundLink, "did not find link in idx=%d %v in %v", tdIdx, link, td.links)
			}
			countUsedIdx[tdIdx]++
			stepCountIdx[tdIdx]++

			existingIdx, ok := usedPoints[p]
			if !ok {
				usedPoints[p] = tdIdx
				for _, np := range nextPoints {
					_, present := usedPoints[np]
					if !present {
						newPoints = append(newPoints, np)
					}
				}
			} else {
				stepConflictCount[p]++
				assert.Equal(t, existingIdx, tdIdx, "conflict on %v step %d ctx %s", p, d, ctx.String())
			}
		}
		stepConflictSummary := make(map[int]int)
		for _, v := range stepConflictCount {
			stepConflictSummary[v]++
		}
		_, impossible := stepConflictSummary[3]
		assert.False(t, impossible)

		Log.Tracef("Run: %2d %4d : %4d %2d %v", d, len(latestPoints), stepConflictSummary[1], stepConflictSummary[2], stepCountIdx)

		newIndexList := make([]int, 0, len(countUsedIdx))
		for trIdx := range countUsedIdx {
			newIndexList = append(newIndexList, trIdx)
		}
		sort.Ints(newIndexList)

		if EqualIntSlice(currentIndexList, newIndexList) {
			stableIndexList++
		} else {
			stableIndexList = 0
			currentIndexList = newIndexList
		}

		if stableIndexList >= verifyStable {
			break
		}

		latestPoints = newPoints
	}
	Log.Debug(ctx.String(), stepStable-verifyStable, currentIndexList)

	return stepStable - verifyStable, currentIndexList
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func EqualIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Stupid reverse engineering of trio index that works for main and non main points
func findTrioIndex(c Point, np [3]Point, ctx *GrowthContext) (int, [3]int) {
	link := [3]int{getTrioIdxNearestMain(c, ctx), getTrioIdxNearestMain(np[1], ctx), getTrioIdxNearestMain(np[2], ctx) }
	toFind := MakeTrioDetails(MakeVector(c, np[0]), MakeVector(c, np[1]), MakeVector(c, np[2]))
	for trIdx, td := range AllTrioDetails {
		if toFind.GetTrio() == td.GetTrio() {
			return trIdx, link
		}
	}
	Log.Errorf("did not find any trio for %v %v %v", c, np, toFind)
	Log.Errorf("All trio index %d %d %d %d", getTrioIdxNearestMain(c, ctx), getTrioIdxNearestMain(np[0], ctx), getTrioIdxNearestMain(np[1], ctx), getTrioIdxNearestMain(np[2], ctx))
	return -1, link
}

func getTrioIdxNearestMain(p Point, ctx *GrowthContext) int {
	return ctx.GetTrioIndex(ctx.GetDivByThree(p.GetNearMainPoint()))
}
