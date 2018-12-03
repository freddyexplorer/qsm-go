package m3space

import "fmt"

type AccessedEventID struct {
	id     EventID
	access TickTime
}

type Node interface {
	IsRoot() bool
	GetLastAccessed() TickTime
	IsActive(space *Space) bool
	HowManyColors(space *Space) uint8
	GetColorMask(space *Space) uint8
	IsEventAlreadyPresent(id EventID) bool
	IsOld(space *Space) bool
	GetStateString() string
}

type ActiveNode struct {
	Pos              Point
	root             bool
	accessedEventIDS []AccessedEventID
	connections      []*Connection
}

type SavedNode struct {
	root             bool
	accessedEventIDS []AccessedEventID
	connections      []int8
}

type Connection struct {
	Id     int8
	P1, P2 Point
}

/***************************************************************/
// Node Functions
/***************************************************************/

func NewNode(p Point) *ActiveNode {
	n := ActiveNode{}
	n.Pos = p
	return &n
}

func (node *ActiveNode) SetRoot(id EventID, time TickTime) {
	node.root = true
	node.accessedEventIDS = make([]AccessedEventID, 1)
	node.accessedEventIDS[0] = AccessedEventID{id, time,}
}

func (node *ActiveNode) HasFreeConnections(space *Space) bool {
	return node.connections == nil || len(node.connections) < space.MaxConnections
}

func (node *ActiveNode) AddConnection(conn *Connection, space *Space) int {
	if !node.HasFreeConnections(space) {
		return -1
	}
	if node.connections == nil {
		node.connections = make([]*Connection, 0, 3)
	}
	index := len(node.connections)
	node.connections = append(node.connections, conn)
	return index
}

func (node *ActiveNode) IsAlreadyConnected(otherNode *ActiveNode) bool {
	for _, conn := range node.connections {
		if conn.IsConnectedTo(otherNode.Pos) {
			return true
		}
	}
	return false
}

func (node *ActiveNode) IsRoot() bool {
	return node.root
}

func (node *ActiveNode) GetLastAccessed() TickTime {
	bestTime := TickTime(0)
	for _, ae := range node.accessedEventIDS {
		if ae.access > bestTime {
			bestTime = ae.access
		}
	}
	return bestTime
}

func (ae AccessedEventID) IsActive(space *Space) bool {
	return Distance(space.currentTime-ae.access) <= space.EventOutgrowthThreshold
}

func (node *ActiveNode) IsActive(space *Space) bool {
	if node.IsRoot() {
		return true
	}
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			return true
		}
	}
	return false
}

func (node *ActiveNode) HowManyColors(space *Space) uint8 {
	return countOnes(node.GetColorMask(space))
}

func countOnes(m uint8) uint8 {
	return ((m >> 7) & 1) + ((m >> 6) & 1) + ((m >> 5) & 1) + ((m >> 4) & 1) + ((m >> 3) & 1) + ((m >> 2) & 1) + ((m >> 1) & 1) + (m & 1)
}

func (node *ActiveNode) GetColorMask(space *Space) uint8 {
	if node.root {
		return uint8(space.events[node.accessedEventIDS[0].id].color)
	}
	m := uint8(0)
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			m |= uint8(space.events[ae.id].color)
		}
	}
	return m
}

func (node *ActiveNode) CanReceiveOutgrowth(newPosEo *NewPossibleOutgrowth) bool {
	if !node.IsEventAlreadyPresent(newPosEo.event.id) {
		return false
	}
	return true
}

func (node *ActiveNode) IsEventAlreadyPresent(id EventID) bool {
	for _, ae := range node.accessedEventIDS {
		if ae.id == id {
			return false
		}
	}
	return true
}

func (node *ActiveNode) AddOutgrowth(id EventID, time TickTime) {
	node.accessedEventIDS = append(node.accessedEventIDS, AccessedEventID{id, time,})
}

func (node *ActiveNode) IsOld(space *Space) bool {
	if node.IsRoot() {
		return false
	}
	return Distance(space.currentTime-node.GetLastAccessed()) >= space.EventOutgrowthOldThreshold
}

func (node *ActiveNode) String() string {
	return fmt.Sprintf("%v:%d:%d", node.Pos, len(node.connections), len(node.accessedEventIDS))
}

func (node *ActiveNode) GetStateString() string {
	connIds := make([]string, len(node.connections))
	for i, conn := range node.connections {
		connIds[i] = AllConnectionsIds[conn.Id].GetName()
	}
	if node.root {
		return fmt.Sprintf("%v: root %v, %v", node.Pos, node.accessedEventIDS, connIds)
	}
	return fmt.Sprintf("%v: %v, %v", node.Pos, node.accessedEventIDS, connIds)
}

/***************************************************************/
// Connection Functions
/***************************************************************/

func (conn *Connection) GetConnId() int8 {
	return conn.Id
}

func (conn *Connection) GetConnectionDetails() ConnectionDetails {
	return AllConnectionsIds[conn.Id]
}

func (conn *Connection) IsConnectedTo(point Point) bool {
	return conn.P1 == point || conn.P2 == point
}

func (conn *Connection) GetColorMask(space *Space) uint8 {
	n1 := space.GetNode(conn.P1)
	n2 := space.GetNode(conn.P2)
	// Connection color mask of all event outgrowth that match
	if n1 != nil && n2 != nil {
		return n1.GetColorMask(space) & n2.GetColorMask(space)
	}
	return uint8(0)
}

func (conn *Connection) HowManyColors(space *Space) uint8 {
	return countOnes(conn.GetColorMask(space))
}

func (conn *Connection) IsOld(space *Space) bool {
	n1 := space.GetNode(conn.P1)
	n2 := space.GetNode(conn.P2)
	if n1 != nil && n2 != nil {
		return n1.IsOld(space) && n2.IsOld(space)
	}
	return false
}

func (space *Space) makeConnection(n1, n2 *ActiveNode) *Connection {
	if !n1.HasFreeConnections(space) {
		Log.Trace("Node 1", n1, "does not have free connections")
		return nil
	}
	if !n2.HasFreeConnections(space) {
		Log.Trace("Node 2", n2, "does not have free connections")
		return nil
	}
	if n1.IsAlreadyConnected(n2) {
		Log.Trace("Connection between 2 points", n1.Pos, n2.Pos, "already connected!")
		return nil
	}

	d := DS(n1.Pos, n2.Pos)
	if !(d == 1 || d == 2 || d == 3 || d == 5) {
		Log.Error("Connection between 2 points", n1.Pos, n2.Pos, "that are not 1, 2, 3 or 5 DS away!")
		return nil
	}
	// All good create connection
	bv := MakeVector(n1.Pos, n2.Pos)
	cd := AllConnectionsPossible[bv]
	c1 := &Connection{cd.GetIntId(), n1.Pos, n2.Pos}
	space.activeConnections = append(space.activeConnections, c1)
	n1done := n1.AddConnection(c1, space)
	c2 := &Connection{-cd.GetIntId(), n2.Pos, n1.Pos}
	n2done := n2.AddConnection(c2, space)
	if n1done < 0 || n2done < 0 {
		Log.Error("Node1 connection association", n1done, "or Node2", n2done, "did not happen!!")
		return nil
	}
	return c1
}

/***************************************************************/
// Saved Node Functions
/***************************************************************/

func (node *SavedNode) IsRoot() bool {
	return node.root
}

func (node *SavedNode) GetLastAccessed() TickTime {
	bestTime := TickTime(0)
	for _, ae := range node.accessedEventIDS {
		if ae.access > bestTime {
			bestTime = ae.access
		}
	}
	return bestTime
}

func (node *SavedNode) IsActive(space *Space) bool {
	if node.IsRoot() {
		return true
	}
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			return true
		}
	}
	return false
}

func (node *SavedNode) HowManyColors(space *Space) uint8 {
	return countOnes(node.GetColorMask(space))
}

func (node *SavedNode) GetColorMask(space *Space) uint8 {
	if node.root {
		return uint8(space.events[node.accessedEventIDS[0].id].color)
	}
	m := uint8(0)
	for _, ae := range node.accessedEventIDS {
		if ae.IsActive(space) {
			m |= uint8(space.events[ae.id].color)
		}
	}
	return m
}

func (node *SavedNode) IsEventAlreadyPresent(id EventID) bool {
	for _, ae := range node.accessedEventIDS {
		if ae.id == id {
			return false
		}
	}
	return true
}

func (node *SavedNode) IsOld(space *Space) bool {
	if node.IsRoot() {
		return false
	}
	return Distance(space.currentTime-node.GetLastAccessed()) >= space.EventOutgrowthOldThreshold
}

func (node *SavedNode) String() string {
	return fmt.Sprintf("%s:%d:%d", "Saved", len(node.connections), len(node.accessedEventIDS))
}

func (node *SavedNode) GetStateString() string {
	connIds := make([]string, len(node.connections))
	for i, connId := range node.connections {
		connIds[i] = AllConnectionsIds[connId].GetName()
	}
	if node.root {
		return fmt.Sprintf("%s: root %v, %v", "Saved", node.accessedEventIDS, connIds)
	}
	return fmt.Sprintf("%s: %v, %v", "Saved", node.accessedEventIDS, connIds)
}
