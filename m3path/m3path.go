package m3path

import (
	"fmt"
	"github.com/freddy33/qsm-go/m3point"
	"github.com/freddy33/qsm-go/m3util"
	"strings"
)

var Log = m3util.NewLogger("m3path", m3util.INFO)

type PathContext struct {
	ctx             *m3point.TrioIndexContext
	rootTrioId      m3point.TrioIndex
	rootPathLinks   [3]*PathLink
	possiblePathIds map[PathIdKey][2]NextPathLink
}

type PathIdKey struct {
	previousMainDivThree uint64
	previousMainTrioId   m3point.TrioIndex
	previousMainConnId   m3point.ConnectionId
	previousTrioId       m3point.TrioIndex
	previousConnId       m3point.ConnectionId
}

type NextPathLink struct {
	connId     m3point.ConnectionId
	nextTrioId m3point.TrioIndex
}

// A single path link between *src* node to one of the next path node *dst* using the connection Id
type PathLink struct {
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	src *PathNode
	// The connection used by the link path
	connId m3point.ConnectionId
	// After travelling the connId the pointer to the next path node
	dst *PathNode
}

// The link graph node of a path, representing one point on the graph
// Points to the 2 path links usable from here
type PathNode struct {
	p m3point.Point
	// Distance from root
	d int
	// From which link this node came from
	from *PathLink
	// The current trio index of the path point
	trioId m3point.TrioIndex
	// After travelling the connId of the above cur.connId there will be 2 new path possible for
	next [2]*PathLink
}

/***************************************************************/
// PathContext Functions
/***************************************************************/

func (pathCtx *PathContext) dumpInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n%s: [", pathCtx.ctx.String(), pathCtx.rootTrioId.String()))
	for i, pl := range pathCtx.rootPathLinks {
		sb.WriteString("\n")
		if pl != nil {
			sb.WriteString(fmt.Sprintf("%d:%s,", i, pl.dumpInfo(0)))
		} else {
			sb.WriteString(fmt.Sprintf("%d:nil,", i))
		}
	}
	sb.WriteString("]")
	return sb.String()
}

/***************************************************************/
// PathLink Functions
/***************************************************************/

func makeRootPathLink(connId m3point.ConnectionId) *PathLink {
	res := PathLink{}
	res.src = nil
	res.connId = connId
	return &res
}

func (pl *PathLink) setDestTrioIdx(p m3point.Point, tdId m3point.TrioIndex) {
	res := PathNode{}
	res.p = p
	if pl.src != nil {
		res.d = pl.src.d + 1
	} else {
		res.d = 1
	}
	res.from = pl
	res.trioId = tdId
	pl.dst = &res
}

func (pl *PathLink) String() string {
	return fmt.Sprintf("PL-%s", pl.connId.String())
}

func (pl *PathLink) dumpInfo(ident int) string {
	var sb strings.Builder
	sb.WriteString(pl.String())
	if pl.dst != nil {
		sb.WriteString(":{")
		sb.WriteString(pl.dst.dumpInfo(ident+1))
		sb.WriteString("}")
	} else {
		sb.WriteString(":{nil}")
	}
	return sb.String()
}

/***************************************************************/
// PathNode Functions
/***************************************************************/

func (pn *PathNode) addPathLinks(connIds... m3point.ConnectionId) {
	if Log.DoAssert() {
		td := m3point.GetTrioDetails(pn.trioId)
		if td == nil {
			Log.Errorf("creating a path link with source node %s pointing to non existent trio index", pn.String())
			return
		}
		if !td.HasConnections(connIds[0], connIds[1]) {
			Log.Errorf("creating a path link with source node %s using connections %v not present in trio", pn.String(), connIds)
			return
		}
	}
	for j := 0; j < 2; j++ {
		res := PathLink{}
		res.src = pn
		res.connId = connIds[j]
		pn.next[j] = &res
	}
}

func (pn *PathNode) String() string {
	return fmt.Sprintf("PN%v-%3d-%s", pn.p, pn.d, pn.trioId.String())
}

func (pn *PathNode) dumpInfo(ident int) string {
	var sb strings.Builder
	sb.WriteString(pn.String())
	if pn.trioId != m3point.NilTrioIndex && (pn.next[0] != nil || pn.next[1] != nil) {
		sb.WriteString("[")
		for i, pl := range pn.next {
			sb.WriteString("\n")
			for k := 0; k < ident; k++ {
				sb.WriteString("  ")
			}
			if pl != nil {
				sb.WriteString(fmt.Sprintf("%d:%s,", i, pl.dumpInfo(ident)))
			} else {
				sb.WriteString(fmt.Sprintf("%d:nil,", i))
			}
		}
		sb.WriteString("]")
	} else {
		sb.WriteString("[]")
	}
	return sb.String()
}











/***************************************************************/
// LEGACY Path Functions
/***************************************************************/


// An element in the path from event base node to latest outgrowth
// Forward is from event to outgrowth
// Backwards is from latest outgrowth to event
type PathElement interface {
	IsEnd() bool
	NbForwardElements() int
	GetForwardConnId(idx int) int8
	GetForwardElement(idx int) PathElement
	Copy() PathElement
	SetLastNext(path PathElement)
	GetLength() int
}

// End of path marker
type EndPathElement int8

// The int8 here is the forward connection Id
type SimplePathElement struct {
	forwardConnId int8
	next          PathElement
}

// We count only forward fork
type ForkPathElement struct {
	simplePaths []*SimplePathElement
}

var TheEnd = EndPathElement(0)

/***************************************************************/
// Simple Path Functions
/***************************************************************/

func (spe EndPathElement) IsEnd() bool {
	return true
}

func (spe EndPathElement) NbForwardElements() int {
	return 0
}

func (spe EndPathElement) GetForwardConnId(idx int) int8 {
	return int8(spe)
}

func (spe EndPathElement) GetForwardElement(idx int) PathElement {
	return nil
}

func (spe EndPathElement) Copy() PathElement {
	return spe
}

func (spe EndPathElement) SetLastNext(path PathElement) {
	Log.Fatalf("cannot set last on end element")
}

func (spe EndPathElement) GetLength() int {
	return 0
}

/***************************************************************/
// Simple Path Functions
/***************************************************************/

func (spe *SimplePathElement) IsEnd() bool {
	return false
}

func (spe *SimplePathElement) NbForwardElements() int {
	return 1
}

func (spe *SimplePathElement) GetForwardConnId(idx int) int8 {
	if idx != 0 {
		Log.Fatalf("index out of bound for %d", idx)
	}
	return spe.forwardConnId
}

func (spe *SimplePathElement) GetForwardElement(idx int) PathElement {
	if idx != 0 {
		Log.Fatalf("index out of bound for %d", idx)
	}
	return spe.next
}

func (spe *SimplePathElement) Copy() PathElement {
	return spe.internalCopy()
}

func (spe *SimplePathElement) internalCopy() *SimplePathElement {
	if spe.next == nil {
		return &SimplePathElement{spe.forwardConnId, nil}
	}
	return &SimplePathElement{spe.forwardConnId, spe.next.Copy()}
}

func (spe *SimplePathElement) SetLastNext(path PathElement) {
	if spe.next == nil {
		spe.next = path
	} else {
		spe.next.SetLastNext(path)
	}
}

func (spe *SimplePathElement) GetLength() int {
	if spe.next == nil {
		return 1
	} else {
		return 1 + spe.next.GetLength()
	}
}

/***************************************************************/
// Forked Path Functions
/***************************************************************/

func (fpe *ForkPathElement) IsEnd() bool {
	return false
}

func (fpe *ForkPathElement) NbForwardElements() int {
	return len(fpe.simplePaths)
}

func (fpe *ForkPathElement) GetForwardConnId(idx int) int8 {
	return fpe.simplePaths[idx].GetForwardConnId(0)
}

func (fpe *ForkPathElement) GetForwardElement(idx int) PathElement {
	return fpe.simplePaths[idx].GetForwardElement(0)
}

func (fpe *ForkPathElement) Copy() PathElement {
	res := ForkPathElement{make([]*SimplePathElement, len(fpe.simplePaths))}
	for i, spe := range fpe.simplePaths {
		res.simplePaths[i] = spe.internalCopy()
	}
	return &res
}

func (fpe *ForkPathElement) SetLastNext(path PathElement) {
	for _, spe := range fpe.simplePaths {
		spe.SetLastNext(path)
	}
}

func (fpe *ForkPathElement) GetLength() int {
	length := fpe.simplePaths[0].GetLength()
	if Log.IsDebug() {
		// All length should be identical
		for i := 1; i < len(fpe.simplePaths); i++ {
			otherLength := fpe.simplePaths[i].GetLength()
			if otherLength != length {
				Log.Errorf("fork points to 2 path with diff length %d != %d", length, otherLength)
			}
		}
	}
	return length
}

/***************************************************************/
// Merge Path Functions
/***************************************************************/

func MergePath(path1, path2 PathElement) PathElement {
	if path1 == nil && path2 == nil {
		return nil
	}
	if (path1 != nil && path2 == nil) || (path1 == nil && path2 != nil) {
		Log.Errorf("cannot merge path if one nil and not the other")
		return nil
	}
	if path1.GetLength() != path2.GetLength() {
		Log.Errorf("cannot merge path of different length")
		return nil
	}
	nb1 := path1.NbForwardElements()
	nb2 := path2.NbForwardElements()
	if nb1 == 1 && nb2 == 1 {
		p1ConnId := path1.GetForwardConnId(0)
		p2ConnId := path2.GetForwardConnId(0)
		p1Next := path1.GetForwardElement(0)
		p2Next := path2.GetForwardElement(0)
		if p1ConnId == p2ConnId {
			return &SimplePathElement{p1ConnId, MergePath(p1Next, p2Next)}
		}
		if p1Next != nil {
			p1Next = p1Next.Copy()
		}
		if p2Next != nil {
			p2Next = p2Next.Copy()
		}
		fpe := ForkPathElement{make([]*SimplePathElement, 2)}
		fpe.simplePaths[0] = &SimplePathElement{p1ConnId, p1Next}
		fpe.simplePaths[1] = &SimplePathElement{p2ConnId, p2Next}
		return &fpe
	}
	pathsPerConnId := make(map[int8][]*SimplePathElement)
	for i := 0; i < nb1; i++ {
		addCopyToMap(path1, i, &pathsPerConnId)
	}
	for i := 0; i < nb2; i++ {
		addCopyToMap(path2, i, &pathsPerConnId)
	}
	i := 0
	res := ForkPathElement{make([]*SimplePathElement, len(pathsPerConnId))}
	for connId, paths := range pathsPerConnId {
		if len(paths) == 1 {
			res.simplePaths[i] = paths[0]
			i++
		} else if len(paths) == 2 {
			res.simplePaths[i] = &SimplePathElement{connId, MergePath(paths[0].GetForwardElement(0), paths[1].GetForwardElement(0))}
			i++
		} else {
			Log.Errorf("Cannot have paths in merge for same connection ids not 1 or 2 for %d %d", connId, len(paths))
		}
	}
	return &res
}

func addCopyToMap(path PathElement, idx int, pathsPerConnId *map[int8][]*SimplePathElement) {
	connId := path.GetForwardConnId(idx)
	next := path.GetForwardElement(idx)
	if next != nil {
		next = next.Copy()
	}
	paths, ok := (*pathsPerConnId)[connId]
	newPath := &SimplePathElement{connId, next}
	if !ok {
		paths = make([]*SimplePathElement, 1)
		paths[0] = newPath
	} else {
		paths = append(paths, newPath)
	}
	(*pathsPerConnId)[connId] = paths
}