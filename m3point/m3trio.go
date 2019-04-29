package m3point

import (
	"fmt"
	"sort"
	"strings"
)

// Trio of connection vectors from any point using connections only
type Trio [3]Point

type TrioIndex uint8

// Keeping track of how base trio connects
type TrioLink struct {
	// The trio index of the source
	a TrioIndex
	// The 2 possible trio index of the destination
	b, c TrioIndex
}

// Defining a list type to manage uniqueness and ordering
type TrioLinkList []*TrioLink

// A bigger struct than Trio to keep more info on how points grow from a trio index
type TrioDetails struct {
	id    TrioIndex
	conns [3]*ConnectionDetails
	Links TrioLinkList
}

// Defining a list type to manage uniqueness and ordering
type TrioDetailList []*TrioDetails

// All the initialized arrays used to navigate the switch between base trio index at base points
var allBaseTrio [8]Trio
var validNextTrio [12][2]TrioIndex
var AllMod4Permutations [12][4]TrioIndex
var AllMod8Permutations [12][8]TrioIndex

const (
	NbTrioDsIndex = 7
	NilTrioIndex  = TrioIndex(255)
)

// All the possible Trio details used
var allTrioDetails TrioDetailList
var allTrioLinks TrioLinkList

/***************************************************************/
// Util Functions
/***************************************************************/

func PosMod2(i uint64) uint64 {
	return i & 0x0000000000000001
}

func PosMod4(i uint64) uint64 {
	return i & 0x0000000000000003
}

func PosMod8(i uint64) uint64 {
	return i & 0x0000000000000007
}

/***************************************************************/
// TrioIndex Functions
/***************************************************************/

func (trIdx TrioIndex) IsBaseTrio() bool {
	return trIdx < 8
}

func (trIdx TrioIndex) String() string {
	return fmt.Sprintf("T%03d", trIdx)
}

// Test if in the base trio index i2 is pointing to the negative trio of i1 T[i1] == T'[i2]
func isPrime(i1, i2 TrioIndex) bool {
	if !i1.IsBaseTrio() || !i2.IsBaseTrio() {
		Log.Errorf("cannot compare prime for non base trio index %d %d")
		return false
	}
	if i1 > i2 {
		return i1 == i2+4
	}
	if i2 > i1 {
		return i2 == i1+4
	}
	return false
}

/***************************************************************/
// TrioLink Functions
/***************************************************************/

func makeTrioLink(a, b, c TrioIndex) TrioLink {
	// The destination should be ordered by smaller first
	if c < b {
		return TrioLink{a, c, b,}
	}
	return TrioLink{a, b, c,}
}

func makeTrioLinkFromInt(a, b, c int) TrioLink {
	return makeTrioLink(TrioIndex(a), TrioIndex(b), TrioIndex(c))
}

func (tl TrioLink) sameBC() bool {
	return tl.b == tl.c
}

func (tl *TrioLink) String() string {
	return fmt.Sprintf("[%d %d %d]", tl.a, tl.b, tl.c)
}

func (tl *TrioLink) Contains(i TrioIndex) bool {
	return tl.a == i || tl.b == i || tl.c == i
}

func (tl *TrioLink) ContainsAB(a, b TrioIndex) bool {
	return tl.a == a && (tl.b == b || tl.c == b)
}

/***************************************************************/
// TrioLinkList Functions
/***************************************************************/

func (l TrioLinkList) Exists(tl *TrioLink) bool {
	present := false
	for _, trL := range l {
		if *trL == *tl {
			present = true
			break
		}
	}
	return present
}

func (l *TrioLinkList) addUnique(tl *TrioLink) bool {
	b := l.Exists(tl)
	if !b {
		*l = append(*l, tl)
	}
	return b
}

func (l *TrioLinkList) addAll(l2 *TrioLinkList) {
	for _, tl := range *l2 {
		l.addUnique(tl)
	}
}

func (l TrioLinkList) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d : ", len(l)))
	for _, tl := range l {
		b.WriteString(tl.String())
		b.WriteString(" ")
	}
	return b.String()
}

func (l TrioLinkList) Len() int {
	return len(l)
}

func (l TrioLinkList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l TrioLinkList) Less(i, j int) bool {
	t1 := l[i]
	t2 := l[j]
	d := t1.a - t2.a
	if d != 0 {
		return d < 0
	}
	d = t1.b - t2.b
	if d != 0 {
		return d < 0
	}
	d = t1.c - t2.c
	if d != 0 {
		return d < 0
	}
	return false
}

/***************************************************************/
// TrioDetailsList Functions
/***************************************************************/

func (l *TrioDetailList) IdList() []TrioIndex {
	res := make([]TrioIndex, len(*l))
	for i, td := range *l {
		res[i] = td.GetId()
	}
	return res
}

func (l *TrioDetailList) ExistsByTrio(tr *TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.GetTrio() == tr.GetTrio() {
			present = true
			break
		}
	}
	return present
}

func (l *TrioDetailList) ExistsById(tr *TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.id == tr.id {
			present = true
			break
		}
	}
	return present
}

func (l *TrioDetailList) addUnique(td *TrioDetails) bool {
	b := l.ExistsByTrio(td)
	if !b {
		*l = append(*l, td)
	}
	return b
}

func (l *TrioDetailList) addWithLinks(td *TrioDetails) bool {
	present := false
	for _, trL := range *l {
		if trL.GetTrio() == td.GetTrio() {
			trL.Links.addAll(&td.Links)
			present = true
			break
		}
	}
	if !present {
		*l = append(*l, td)
	}
	return present
}

func (l TrioDetailList) Len() int {
	return len(l)
}

func (l TrioDetailList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l TrioDetailList) Less(i, j int) bool {
	tr1 := l[i]
	tr2 := l[j]
	ds1Index := tr1.GetDSIndex()
	diffDS := ds1Index - tr2.GetDSIndex()

	// Order by ds index first
	if diffDS < 0 {
		return true
	} else if diffDS > 0 {
		return false
	} else {
		// Same ds index
		if ds1Index == 0 {
			// Base trio, keep order as defined with 0-4 prime -> 5-7
			var k, l int
			for bi, bt := range allBaseTrio {
				if bt == tr1.GetTrio() {
					k = bi
				}
				if bt == tr2.GetTrio() {
					l = bi
				}
			}
			return k < l
		} else {
			// order by conn id, first ABS number, then pos > neg
			for k, cd1 := range tr1.conns {
				cd2 := tr2.conns[k]
				if cd1.GetId() != cd2.GetId() {
					return IsLessConnId(cd1, cd2)
				}
			}
		}
	}
	Log.Errorf("Should not get here for %v compare to %v", *tr1, *tr2)
	return false
}

/***************************************************************/
// Init Functions
// TODO: Find a better way than Init
/***************************************************************/

func init() {
	// Initial Trio 0
	allBaseTrio[0] = MakeBaseConnectingVectorsTrio([3]Point{{1, 1, 0}, {-1, 0, -1}, {0, -1, 1}})
	for i := 1; i < 4; i++ {
		allBaseTrio[i] = allBaseTrio[i-1].PlusX()
	}
	// Initial Trio 0 prime
	for i := 0; i < 4; i++ {
		allBaseTrio[i+4] = allBaseTrio[i].Neg()
	}

	initValidTrios()
	initMod4Permutations()
	initMod8Permutations()
	initConnectionDetails()
	fillAllTrioDetails()
}

func initValidTrios() {
	// Valid next trio are all but prime
	idx := 0
	for i := TrioIndex(0); i < 4; i++ {
		for j := TrioIndex(4); j < 8; j++ {
			// j index cannot be the prime (neg) trio
			if !isPrime(i, j) {
				validNextTrio[idx] = [2]TrioIndex{i, j}
				idx++
			}
		}
	}
}

func initMod4Permutations() {
	p := TrioIndexPermBuilder{4, 0, make([][]TrioIndex, 12)}
	p.fill(0, make([]TrioIndex, p.size))
	for pIdx := 0; pIdx < len(AllMod4Permutations); pIdx++ {
		for i := 0; i < 4; i++ {
			AllMod4Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

func initMod8Permutations() {
	p := TrioIndexPermBuilder{8, 0, make([][]TrioIndex, 12)}
	// In 8 size permutation the first index always 0 since we use all the indexes
	first := make([]TrioIndex, p.size)
	first[0] = TrioIndex(0)
	p.fill(1, first)
	for pIdx := 0; pIdx < len(AllMod8Permutations); pIdx++ {
		for i := 0; i < 8; i++ {
			AllMod8Permutations[pIdx][i] = p.collector[pIdx][i]
		}
	}
}

/***************************************************************/
// Trio Functions
/***************************************************************/
var reverse3Map = [3]TrioIndex{2, 1, 0}

func GetNumberOfBaseTrio() int {
	return len(allBaseTrio)
}

func GetBaseTrio(trioIdx TrioIndex) Trio {
	return allBaseTrio[trioIdx]
}

func GetValidNextTrioPair(nextValidIdx TrioIndex) [2]TrioIndex {
	return validNextTrio[nextValidIdx]
}

func MakeBaseConnectingVectorsTrio(points [3]Point) Trio {
	res := Trio{}
	// All points should be a connecting vector
	for _, p := range points {
		if !p.IsBaseConnectingVector() {
			Log.Error("Trying to create a base trio out of non base vector!", p)
			return res
		}
	}
	for _, p := range points {
		for i := 0; i < 3; i++ {
			if p[i] == 0 {
				res[reverse3Map[i]] = p
			}
		}
	}
	return res
}

func (t Trio) GetDSIndex() int {
	if t[0].DistanceSquared() == int64(1) {
		return 1
	} else {
		switch t[1].DistanceSquared() {
		case int64(2):
			return 0
		case int64(3):
			return 2
		case int64(5):
			return 3
		}
	}
	Log.Errorf("Did not find correct index for %v", t)
	return -1
}

func (t Trio) PlusX() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].PlusX(), t[1].PlusX(), t[2].PlusX()})
}

func (t Trio) Neg() Trio {
	return MakeBaseConnectingVectorsTrio([3]Point{t[0].Neg(), t[1].Neg(), t[2].Neg()})
}

// Return the 6 connections possible +X, -X, +Y, -Y, +Z, -Z vectors between 2 Trio
func GetNonBaseConnections(tA, tB Trio) [6]Point {
	res := [6]Point{}
	for _, aVec := range tA {
		switch aVec.X() {
		case 0:
			// Nothing connect
		case 1:
			// This is +X, find the -X on the other side
			bVec := tB.getMinusXVector()
			res[0] = XFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -X, find the +X on the other side
			bVec := tB.getPlusXVector()
			res[1] = XFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Y() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Y, find the -Y on the other side
			bVec := tB.getMinusYVector()
			res[2] = YFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Y, find the +Y on the other side
			bVec := tB.getPlusYVector()
			res[3] = YFirst.Neg().Add(bVec).Sub(aVec)
		}
		switch aVec.Z() {
		case 0:
			// Nothing connect
		case 1:
			// This is +Z, find the -Z on the other side
			bVec := tB.getMinusZVector()
			res[4] = ZFirst.Add(bVec).Sub(aVec)
		case -1:
			// This is -Z, find the +Z on the other side
			bVec := tB.getPlusZVector()
			res[5] = ZFirst.Neg().Add(bVec).Sub(aVec)
		}
	}
	return res
}

func (t Trio) getPlusXVector() Point {
	for _, vec := range t {
		if vec.X() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +X vector")
	return Origin
}

func (t Trio) getMinusXVector() Point {
	for _, vec := range t {
		if vec.X() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -X vector")
	return Origin
}

func (t Trio) getPlusYVector() Point {
	for _, vec := range t {
		if vec.Y() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +Y vector")
	return Origin
}

func (t Trio) getMinusYVector() Point {
	for _, vec := range t {
		if vec.Y() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -Y vector")
	return Origin
}

func (t Trio) getPlusZVector() Point {
	for _, vec := range t {
		if vec.Z() == 1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a +Z vector")
	return Origin
}

func (t Trio) getMinusZVector() Point {
	for _, vec := range t {
		if vec.Z() == -1 {
			return vec
		}
	}
	Log.Error("Impossible! For all base trio there should be a -Z vector")
	return Origin
}

/***************************************************************/
// TrioDetails Functions
/***************************************************************/
var _traceCounter = 0

func GetNumberOfTrioDetails() TrioIndex {
	return TrioIndex(len(allTrioDetails))
}

func GetTrioDetails(trIdx TrioIndex) *TrioDetails {
	return allTrioDetails[trIdx]
}

func MakeTrioDetails(points ...Point) *TrioDetails {
	// All points should be a connection details
	cds := make([]*ConnectionDetails, 3)
	for i, p := range points {
		cd, ok := allConnectionsByVector[p]
		if !ok {
			Log.Fatalf("trying to create trio with vector not a connection %v", p)
		} else {
			cds[i] = cd
		}
	}
	// Order based on connection details index, and if same index Pos > Neg
	sort.Slice(cds, func(i, j int) bool {
		absDiff := cds[i].GetPosId() - cds[j].GetPosId()
		if absDiff == 0 {
			return cds[i].Id > 0
		}
		return absDiff < 0
	})
	res := TrioDetails{}
	res.id = NilTrioIndex
	for i, cd := range cds {
		res.conns[i] = cd
	}
	if Log.IsTrace() {
		fmt.Println(_traceCounter, res.conns[0].String(), res.conns[1].String(), res.conns[2].String())
		_traceCounter++
	}
	return &res
}

func (td *TrioDetails) String() string {
	return fmt.Sprintf("T%02d: (%s, %s, %s)", td.id, td.conns[0].String(), td.conns[1].String(), td.conns[2].String())
}

func (td *TrioDetails) HasConnection(connId ConnectionId) bool {
	for _, c := range td.conns {
		if c.Id == connId {
			return true
		}
	}
	return false
}

// The passed connId is where come from so is neg in here
func (td *TrioDetails) OtherConnectionsFrom(connId ConnectionId) [2]*ConnectionDetails {
	res := [2]*ConnectionDetails{nil, nil}
	idx := 0

	if td.HasConnection(-connId) {
		for _, c := range td.conns {
			if c.Id != -connId {
				res[idx] = c
				idx++
			}
		}
	} else {
		Log.Errorf("connection %s is not part of %s and cannot return other connections", connId.String(), td.String())
	}

	return res
}

func (td *TrioDetails) LastOtherConnection(cIds ...ConnectionId) *ConnectionDetails {
	for _, c := range td.conns {
		for _, cId := range cIds {
			if c.Id != cId {
				return c
			}
		}
	}
	return nil
}

func (td *TrioDetails) HasConnections(cIds ...ConnectionId) bool {
	for _, cId := range cIds {
		if !td.HasConnection(cId) {
			return false
		}
	}
	return true
}

func (td *TrioDetails) HasLinkWith(a, b TrioIndex) bool {
	for _, l := range td.Links {
		if l.ContainsAB(a, b) {
			return true
		}
	}
	return false
}

func (td *TrioDetails) GetTrio() Trio {
	return Trio{td.conns[0].Vector, td.conns[1].Vector, td.conns[2].Vector}
}

func (td *TrioDetails) GetConnections() [3]*ConnectionDetails {
	return td.conns
}

func (td *TrioDetails) GetId() TrioIndex {
	return td.id
}

func (td *TrioDetails) IsBaseTrio() bool {
	return td.id < 8
}

func (td *TrioDetails) findConn(vecName string, toFind ...ConnectionId) *ConnectionDetails {
	if !td.IsBaseTrio() {
		Log.Errorf("cannot look for %s conn on non base trio %s", vecName, td.String())
		return nil
	}
	if Log.DoAssert() {
		// verify only one found
		count := 0
		var res *ConnectionDetails
		for _, c := range td.conns {
			for _, ct := range toFind {
				if c.Id == ct {
					res = c
					count++
				}
			}
		}
		if count == 0 {
			Log.Errorf("Impossible! Did not find %s vector using %v in base trio %s", vecName, toFind, td.String())
			return nil
		} else if count > 1 {
			Log.Errorf("Found %d which is more than one %s vector using %v in base trio %s", count, vecName, toFind, td.String())
			return nil
		}
		return res
	} else {
		for _, c := range td.conns {
			for _, ct := range toFind {
				if c.Id == ct {
					return c
				}
			}
		}
		Log.Errorf("Impossible! Did not find %s vector using %v in base trio %s", vecName, toFind, td.String())
		return nil
	}
}

func (td *TrioDetails) getPlusXConn() *ConnectionDetails {
	return td.findConn("+X", 4, 5, 6, 7)
}

func (td *TrioDetails) getMinusXConn() *ConnectionDetails {
	return td.findConn("-X", -4, -5, -6, -7)
}

func (td *TrioDetails) getPlusYConn() *ConnectionDetails {
	return td.findConn("+Y", 4, -5, 8, 9)
}

func (td *TrioDetails) getMinusYConn() *ConnectionDetails {
	return td.findConn("-Y", -4, 5, -8, -9)
}

func (td *TrioDetails) getPlusZConn() *ConnectionDetails {
	return td.findConn("+Z", 6, -7, 8, -9)
}

func (td *TrioDetails) getMinusZConn() *ConnectionDetails {
	return td.findConn("-Z", -6, 7, -8, 9)
}

func (td *TrioDetails) GetDSIndex() int {
	if td.conns[0].DistanceSquared() == int64(1) {
		switch td.conns[1].DistanceSquared() {
		case int64(1):
			return 1
		case int64(2):
			switch td.conns[2].DistanceSquared() {
			case int64(3):
				return 2
			case int64(5):
				return 3
			}
		}
	} else {
		switch td.conns[1].DistanceSquared() {
		case int64(2):
			return 0
		case int64(3):
			switch td.conns[2].DistanceSquared() {
			case int64(3):
				return 4
			case int64(5):
				return 5
			}
		case int64(5):
			return 6
		}
	}
	Log.Errorf("Did not find correct index for %v", *td)
	return -1
}

func fillAllTrioDetails() {
	localTrioLinks := TrioLinkList{}
	localTrioDetails := TrioDetailList{}
	// All base trio first
	for i, tr := range allBaseTrio {
		td := MakeTrioDetails(tr[0], tr[1], tr[2])
		td.id = TrioIndex(i)
		localTrioDetails.addUnique(td)
	}
	// Going through all Trio and all combination of Trio, to find middle points and create new Trios
	for a, tA := range allBaseTrio {
		for b, tB := range allBaseTrio {
			for c, tC := range allBaseTrio {
				thisTrio := makeTrioLinkFromInt(a, b, c)
				alreadyDone := localTrioLinks.addUnique(&thisTrio)
				if !alreadyDone {
					for _, nextTrio := range getNextTriosDetails(tA, tB, tC) {
						nextTrio.Links.addUnique(&thisTrio)
						localTrioDetails.addWithLinks(nextTrio)
					}
				}
			}
		}
	}

	sort.Sort(localTrioDetails)

	// Process all the trio details now that order final
	for i, td := range localTrioDetails {
		trIdx := TrioIndex(i)
		// For all base trio different process
		if i < len(allBaseTrio) {
			// The id should already be set correctly
			if td.id != trIdx {
				Log.Fatalf("incorrect Id for base trio details %v at %d", *td, i)
			}
			// For all base trio add all links containing them
			for j, tl := range localTrioLinks {
				if tl.Contains(trIdx) {
					td.Links.addUnique(localTrioLinks[j])
				}
			}
		} else {
			// The id should not have been set. Adding it now
			if td.id != NilTrioIndex {
				Log.Fatalf("incorrect Id for non base trio details %v at %d", *td, i)
			}
			td.id = trIdx
		}

		// Order the links array
		sort.Sort(td.Links)
	}
	allTrioDetails = localTrioDetails
	allTrioLinks = localTrioLinks
}

// Return the new Trio out of Origin + tA (with next tB or tB/tC)
func getNextTriosDetails(tA, tB, tC Trio) []*TrioDetails {
	// 0 z=0 for first element, x connector, y connector
	// 1 y=0 for first element, x connector, z connector
	// 2 x=0 for first element, y connector, z connector
	res := TrioDetailList{}

	sameBC := tB == tC
	noZ := tA[0]
	var xConnB, yConnB, zConnB Point
	var xConnC, yConnC, zConnC Point
	if noZ.X() > 0 {
		xConnB = MakeVector(noZ, XFirst.Add(tB.getMinusXVector()))
		if !sameBC {
			xConnC = MakeVector(noZ, XFirst.Add(tC.getMinusXVector()))
		}
	} else {
		xConnB = MakeVector(noZ, XFirst.Neg().Add(tB.getPlusXVector()))
		if !sameBC {
			xConnC = MakeVector(noZ, XFirst.Neg().Add(tC.getPlusXVector()))
		}
	}
	if noZ.Y() > 0 {
		yConnB = MakeVector(noZ, YFirst.Add(tB.getMinusYVector()))
		if !sameBC {
			yConnC = MakeVector(noZ, YFirst.Add(tC.getMinusYVector()))
		}
	} else {
		yConnB = MakeVector(noZ, YFirst.Neg().Add(tB.getPlusYVector()))
		if !sameBC {
			yConnC = MakeVector(noZ, YFirst.Neg().Add(tC.getPlusYVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnB, yConnB))
	} else {
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnB, yConnC))
		res.addUnique(MakeTrioDetails(noZ.Neg(), xConnC, yConnB))
	}

	noY := tA[1]
	if noY.X() > 0 {
		xConnB = MakeVector(noY, XFirst.Add(tB.getMinusXVector()))
		if !sameBC {
			xConnC = MakeVector(noY, XFirst.Add(tC.getMinusXVector()))
		}
	} else {
		xConnB = MakeVector(noY, XFirst.Neg().Add(tB.getPlusXVector()))
		if !sameBC {
			xConnC = MakeVector(noY, XFirst.Neg().Add(tC.getPlusXVector()))
		}
	}
	if noY.Z() > 0 {
		zConnB = MakeVector(noY, ZFirst.Add(tB.getMinusZVector()))
		if !sameBC {
			zConnC = MakeVector(noY, ZFirst.Add(tC.getMinusZVector()))
		}
	} else {
		zConnB = MakeVector(noY, ZFirst.Neg().Add(tB.getPlusZVector()))
		if !sameBC {
			zConnC = MakeVector(noY, ZFirst.Neg().Add(tC.getPlusZVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnB, zConnB))
	} else {
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnB, zConnC))
		res.addUnique(MakeTrioDetails(noY.Neg(), xConnC, zConnB))
	}

	noX := tA[2]
	if noX.Y() > 0 {
		yConnB = MakeVector(noX, YFirst.Add(tB.getMinusYVector()))
		if !sameBC {
			yConnC = MakeVector(noX, YFirst.Add(tC.getMinusYVector()))
		}
	} else {
		yConnB = MakeVector(noX, YFirst.Neg().Add(tB.getPlusYVector()))
		if !sameBC {
			yConnC = MakeVector(noX, YFirst.Neg().Add(tC.getPlusYVector()))
		}
	}
	if noX.Z() > 0 {
		zConnB = MakeVector(noX, ZFirst.Add(tB.getMinusZVector()))
		if !sameBC {
			zConnC = MakeVector(noX, ZFirst.Add(tC.getMinusZVector()))
		}
	} else {
		zConnB = MakeVector(noX, ZFirst.Neg().Add(tB.getPlusZVector()))
		if !sameBC {
			zConnC = MakeVector(noX, ZFirst.Neg().Add(tC.getPlusZVector()))
		}
	}
	if sameBC {
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnB, zConnB))
	} else {
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnB, zConnC))
		res.addUnique(MakeTrioDetails(noX.Neg(), yConnC, zConnB))
	}

	return res
}
