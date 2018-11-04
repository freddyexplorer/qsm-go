package m3space

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSingleRedEvent(t *testing.T) {
	DEBUG = true
	SpaceObj.Clear()
	InitConnectionDetails()
	assertEmptySpace(t)

	// Only latest counting
	DrawSelector.EventOutgrowthThreshold = Distance(0)

	SpaceObj.max = 3 * 9
	expectedTime := TickTime(0)
	nbNodes := 1
	nbConnections := 0

	assert.Equal(t, int64(3*9), SpaceObj.max)

	SpaceObj.CreateSingleEventCenter()
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, 1)

	SpaceObj.ForwardTime()
	expectedTime++
	newNodes := 3
	nbNodes += newNodes
	nbConnections += newNodes
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, 3+1)

	SpaceObj.ForwardTime()
	expectedTime++
	newNodes *= 2
	nbNodes += newNodes
	nbConnections += newNodes
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, 6+1)

	SpaceObj.ForwardTime()
	expectedTime++
	newNodes *= 2
	nbNodes += newNodes
	nbConnections += newNodes
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, 12+1)

	SpaceObj.ForwardTime()
	expectedTime++

	newNodes *= 2
	newNodes -= 2 // 2 nodes already done
	newConnections := newNodes

	nbNodes += newNodes
	nbConnections += newConnections
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, 24-2+1)

	SpaceObj.ForwardTime()
	expectedTime++

	newNodes *= 2
	newNodes -= 10 // 10 nodes already done
	newConnections = newNodes

	nbNodes += newNodes
	nbConnections += newConnections
	assertSpaceSingleEvent(t, expectedTime, nbNodes, nbConnections, newNodes+1)

	assertNearMainPoints(t)
}

func assertEmptySpace(t *testing.T) {
	assert.Equal(t, int64(9), SpaceObj.max)
	assert.Equal(t, 0, len(SpaceObj.nodesMap))
	assert.Equal(t, 0, len(SpaceObj.connections))
	assert.Equal(t, 0, len(SpaceObj.events))
	assert.Equal(t, 0, len(SpaceObj.Elements))
}

func assertSpaceSingleEvent(t *testing.T, time TickTime, nbNodes, nbConnections int, nbActive int) {
	assert.Equal(t, time, SpaceObj.currentTime)
	assert.Equal(t, nbNodes, len(SpaceObj.nodesMap))
	assert.Equal(t, nbConnections, len(SpaceObj.connections))
	assert.Equal(t, 1, len(SpaceObj.events))
	assert.Equal(t, nbNodes+nbConnections+6, len(SpaceObj.Elements))
	collectActiveElements := make([]*NodeDrawingElement, 0, 20)
	for _, draw := range SpaceObj.Elements {
		if draw.Key() == NodeActive {
			nodeDrawing, ok := draw.(*NodeDrawingElement)
			assert.True(t, ok, "Node draw element should be of type NodeDrawingElement not %v", draw)
			collectActiveElements = append(collectActiveElements, nodeDrawing)
		}
	}
	assert.Equal(t, nbActive, len(collectActiveElements))
	for _, nodeDraw := range collectActiveElements {
		assert.Equal(t, uint8(1), nodeDraw.sdc.howManyColors())
	}
}

func assertNearMainPoints(t *testing.T) {
	for _, node := range SpaceObj.nodesMap {
		// Find main point attached to node
		var mainPointNode *Node
		if node.point.IsMainPoint() {
			mainPointNode = node
		} else {
			for _, conn := range node.connections {
				if conn != nil {
					if conn.N1.point.IsMainPoint() {
						mainPointNode = conn.N1
						break
					}
					if conn.N2.point.IsMainPoint() {
						mainPointNode = conn.N2
						break
					}
				}
			}
		}
		if mainPointNode != nil {
			assert.Equal(t, node.point.getNearMainPoint(), *(mainPointNode.point))
		}
	}
}

func TestSpace(t *testing.T) {
	DEBUG = true

	SpaceObj.Clear()
	assertEmptySpace(t)

	SpaceObj.SetMax(3)
	SpaceObj.CreatePyramid(1)
	assert.Equal(t, int64(3), SpaceObj.max)
	// Big nodes = (center + center face + middle edge + corner) * (main + 3)
	nbNodes := (1 + 6 + 12 + 8) * 4

	// SKIPPING FOR NOW
	t.Skip("Skipping full space test until all events tested correctly")

	/*******************  STEP 0 ******************/
	nbOutgrowthsStep0 := 4
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(0), SpaceObj.currentTime)

	assertOutgrowth(t, 4)
	assertOutgrowthDistance(t, map[EventID]int{0: 1, 1: 1, 2: 1, 3: 1})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - 4), 1: 4, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - 4), 1: 4, 10: 4})

	/*******************  STEP 1 ******************/
	SpaceObj.ForwardTime()
	// Same elements just color changes
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(1), SpaceObj.currentTime)
	newOutgrowthsStep1 := 4 * 3
	nbOutgrowthsStep1 := nbOutgrowthsStep0 + newOutgrowthsStep1

	assertOutgrowth(t, nbOutgrowthsStep1)
	assertOutgrowthDistance(t, map[EventID]int{0: 3, 1: 3, 2: 3, 3: 3})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep1 - 4), 1: newOutgrowthsStep1 + 4, 10: 4})
	assertOutgrowthColors(t, 1, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep1), 1: nbOutgrowthsStep1, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep1), 1: nbOutgrowthsStep1, 10: 4})

	/*******************  STEP 2 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(2), SpaceObj.currentTime)
	newOutgrowthsStep2 := (4 * 3) + 2
	nbOutgrowthsStep2 := nbOutgrowthsStep1 + newOutgrowthsStep2

	assertOutgrowth(t, nbOutgrowthsStep1+newOutgrowthsStep2)
	assertOutgrowthDistance(t, map[EventID]int{0: 3, 1: 3, 2: 3, 3: 5})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep2 - 4), 1: newOutgrowthsStep2 + 4, 10: 4})
	assertOutgrowthColors(t, 1, map[uint8]int{0: int(nbNodes - (newOutgrowthsStep1 + newOutgrowthsStep2) - 4), 1: newOutgrowthsStep1 + newOutgrowthsStep2 + 4, 10: 4})
	assertOutgrowthColors(t, 2, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep2), 1: nbOutgrowthsStep2, 10: 4})
	assertOutgrowthColors(t, 20, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep2), 1: nbOutgrowthsStep2, 10: 4})

	/*******************  STEP 3 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(3), SpaceObj.currentTime)
	newOutgrowthsStep3 := (4*3)*2 - 1
	nbOutgrowthsStep3 := nbOutgrowthsStep2 + newOutgrowthsStep3
	nb2colorsStep3 := 2

	assertOutgrowth(t, nbOutgrowthsStep3)
	assertOutgrowthDistance(t, map[EventID]int{0: 4, 1: 6, 2: 4, 3: 9})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep3 + nb2colorsStep3 - 4), 1: newOutgrowthsStep3 + 4 - 2*nb2colorsStep3, 2: nb2colorsStep3, 10: 4})
	assertOutgrowthColors(t, 3, map[uint8]int{0: int(nbNodes - nbOutgrowthsStep3 + nb2colorsStep3), 1: nbOutgrowthsStep3 - 2*nb2colorsStep3, 2: nb2colorsStep3, 10: 4})

	/*******************  STEP 4 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(4), SpaceObj.currentTime)
	newOutgrowthsStep4 := (4*3)*4 - 5
	nbOutgrowthsStep4 := nbOutgrowthsStep3 + newOutgrowthsStep4
	nb2colorsStep4 := 5 + nb2colorsStep3

	assertOutgrowth(t, nbOutgrowthsStep4)
	assertOutgrowthDistance(t, map[EventID]int{0: 7, 1: 12, 2: 7, 3: 17})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep4 + nb2colorsStep4 - 4), 1: newOutgrowthsStep4 - 2*nb2colorsStep4 + 4, 2: nb2colorsStep4, 10: 4})

	/*******************  STEP 5 ******************/
	SpaceObj.ForwardTime()
	assertSpaceMax3WithEvents(t, nbNodes)

	assert.Equal(t, TickTime(5), SpaceObj.currentTime)
	newOutgrowthsStep5 := (4*3)*4 - 4
	nbOutgrowthsStep5 := nbOutgrowthsStep4 + newOutgrowthsStep5
	nb2colorsStep5 := 6

	assertOutgrowth(t, nbOutgrowthsStep5)
	assertOutgrowthDistance(t, map[EventID]int{0: 6, 1: 13, 2: 8, 3: 17})
	assertOutgrowthColors(t, 0, map[uint8]int{0: int(nbNodes - newOutgrowthsStep5 + nb2colorsStep5 + 2*2 - 4), 1: newOutgrowthsStep5 - 2*nb2colorsStep5 - 3*2 + 4, 2: nb2colorsStep5, 3: 2, 10: 4})
}

func assertSpaceMax3WithEvents(t *testing.T, nbNodes int) {
	assert.Equal(t, nbNodes, len(SpaceObj.nodesMap))
	assert.Equal(t, nbNodes+35, len(SpaceObj.connections))
	assert.Equal(t, 4, len(SpaceObj.events))
	assert.Equal(t, 2*nbNodes+35+6, len(SpaceObj.Elements))
}

func assertOutgrowth(t *testing.T, expect int) {
	nbOutgrowth := 0
	for _, evt := range SpaceObj.events {
		nbOutgrowth += len(evt.outgrowths)
	}
	assert.Equal(t, expect, nbOutgrowth)
	nbOutgrowth = 0
	for _, node := range SpaceObj.nodesMap {
		nbOutgrowth += len(node.outgrowths)
	}
	assert.Equal(t, expect, nbOutgrowth)
}

func assertOutgrowthDistance(t *testing.T, topOnes map[EventID]int) {
	for _, evt := range SpaceObj.events {
		nbTopOnes := 0
		for _, eo := range evt.outgrowths {
			if eo.distance == Distance(SpaceObj.currentTime-evt.created) {
				assert.Equal(t, eo.state, EventOutgrowthLatest, "Event outgrowth state test failed for evtID=%d node=%v . Should be latest", evt.id, *(eo.node))
				nbTopOnes++
			} else {
				assert.Equal(t, eo.state, EventOutgrowthOld, "Event outgrowth state test failed for evtID=%d node=%v . Should be old", evt.id, *(eo.node))
			}
		}
		assert.Equal(t, topOnes[evt.id], nbTopOnes, "NB top ones expected failed for evtID=%d", evt.id)
	}
}

func assertOutgrowthColors(t *testing.T, threshold Distance, multiOutgrowths map[uint8]int) {
	// map of how many nodes have 0, 1, 2, 3, 4 event outgrowth, the key 10 is for the amount of root
	count := make(map[uint8]int)
	for _, node := range SpaceObj.nodesMap {
		if node.IsRoot() {
			count[10]++
		}
		count[node.HowManyColors(threshold)]++
	}
	for k, v := range count {
		assert.Equal(t, multiOutgrowths[k], v, "color outgrowth count failed for k=%d and th=%d", k, threshold)
	}
}
