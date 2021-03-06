package m3space

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3db"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"sync"
)

var Log = m3util.NewLogger("m3space", m3util.INFO)

type SpaceVisitor interface {
	VisitNode(space *Space, node Node)
	VisitLink(space *Space, srcPoint m3point.Point, connId m3point.ConnectionId)
}

type Space struct {
	env *m3db.QsmEnvironment

	// the int value of the next event id created
	lastIdCounter EventID
	maxEvents int

	// The slice of events where the index is the EventID
	events []*Event

	// The current time of space time
	currentTime DistAndTime

	// The single big map of all the points
	nbNodes int
	nodesMap sync.Map
	// Extracted list from the above map of the current state at currentTime
	latestNodes NodeList
	activeNodes NodeList
	activeLinks NodeLinkList

	nbDeadNodes int

	// Max absolute coordinate in all nodes
	Max m3point.CInt
	// Max number of connections per node
	MaxConnections int
	// Cancel on same event conflict
	blockOnSameEvent int
	// DistAndTime from latest below which to consider event outgrowth active
	EventOutgrowthThreshold DistAndTime
	// DistAndTime from latest above which to consider event outgrowth old
	EventOutgrowthOldThreshold DistAndTime
	// DistAndTime from latest above which to consider event outgrowth dead
	EventOutgrowthDeadThreshold DistAndTime
}

func MakeSpace(env *m3db.QsmEnvironment, max m3point.CInt) Space {
	space := Space{}
	space.env = env
	space.lastIdCounter = 1
	space.maxEvents = 12
	space.events = make([]*Event, space.maxEvents)
	space.currentTime = 0
	space.latestNodes = make([]Node, 0, 1)
	space.activeNodes = make([]Node, 0, 1)
	space.activeLinks = make([]NodeLink, 0, 500)

	space.nbDeadNodes = 0
	space.Max = max
	space.MaxConnections = 3
	space.blockOnSameEvent = 3
	space.SetEventOutgrowthThreshold(DistAndTime(1))
	return space
}

func (space *Space) SetEventOutgrowthThreshold(threshold DistAndTime) {
	if threshold > 2^50 {
		threshold = 0
	}
	space.EventOutgrowthThreshold = threshold
	// Everything more than 3*3 above threshold move to active => old
	space.EventOutgrowthOldThreshold = threshold + 3
	// Everything more than 3*3*3 above threshold move to old => dead
	space.EventOutgrowthDeadThreshold = threshold + 3*3
}

func (space *Space) GetEnv() *m3db.QsmEnvironment {
	return space.env
}

func (space *Space) GetPointPackData() *m3point.PointPackData {
	return m3point.GetPointPackData(space.GetEnv())
}

func (space *Space) GetCurrentTime() DistAndTime {
	return space.currentTime
}

func (space *Space) GetNbActiveNodes() int {
	return len(space.activeNodes)
}

func (space *Space) GetNbNodes() int {
	return space.nbNodes
}

func (space *Space) GetNbActiveLinks() int {
	return len(space.activeLinks)
}

func (space *Space) GetNbEvents() int {
	res := 0
	for _, evt := range space.events {
		if evt != nil {
			res++
		}
	}
	return res
}

func (space *Space) GetEvent(id EventID) *Event {
	return space.events[id]
}

func (space *Space) VisitAll(visitor SpaceVisitor, onlyActive bool) {
	if onlyActive {
		for _, n := range space.activeNodes {
			visitor.VisitNode(space, n)
			nll := n.GetActiveLinks(space)
			for _, nl := range nll {
				visitor.VisitLink(space, nl.GetSrc(), nl.GetConnId())
			}
		}
	} else {
		space.nodesMap.Range(func(pI, nI interface{}) bool {
			visitor.VisitNode(space, nI.(Node))
			return true
		})
	}
}

func (space *Space) CreateSingleEventCenter() *Event {
	return space.CreateEventFromColor(m3point.Origin, RedEvent)
}

func (space *Space) CreatePyramid(pyramidSize m3point.CInt) {
	space.CreateEventFromColor(m3point.Point{3, 0, 3}.Mul(pyramidSize), RedEvent)
	space.CreateEventFromColor(m3point.Point{-3, 3, 3}.Mul(pyramidSize), GreenEvent)
	space.CreateEventFromColor(m3point.Point{-3, -3, 3}.Mul(pyramidSize), BlueEvent)
	space.CreateEventFromColor(m3point.Point{0, 0, -3}.Mul(pyramidSize), YellowEvent)
}

func (space *Space) GetNode(p m3point.Point) Node {
	res, ok := space.nodesMap.Load(p)
	if !ok {
		return nil
	}
	return res.(Node)
}

func (space *Space) newEmptyNode(p m3point.Point) Node {
	an := new(BaseNode)
	an.p = p
	an.head = nil
	return an
}

func (space *Space) getOrCreateNode(p m3point.Point) Node {
	res, loaded := space.nodesMap.LoadOrStore(p, space.newEmptyNode(p))
	if !loaded {
		space.nbNodes++
		for _, c := range p {
			if c > 0 && space.Max < c {
				space.Max = c
			}
			if c < 0 && space.Max < -c {
				space.Max = -c
			}
		}
	}
	return res.(Node)
}

func (space *Space) DisplayState() {
	fmt.Println("========= Space State =========")
	fmt.Println("Current Time", space.currentTime)
	fmt.Println("Nb Nodes", space.GetNbNodes(), "Nb Active Nodes", len(space.activeNodes), ", Nb Connections", len(space.activeLinks), ", Nb Events", space.GetNbEvents())
}
