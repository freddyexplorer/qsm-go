package m3space

import "fmt"

type NewPossibleOutgrowth struct {
	pos      Point
	event    *Event
	from     *EventOutgrowth
	distance Distance
	state    EventOutgrowthState
}

type EventAlreadyGrewThereError struct {
	id  EventID
	pos Point
}

func (e *EventAlreadyGrewThereError) Error() string {
	return fmt.Sprintf("event with id %d already has outgrowth at %v", e.id, e.pos)
}

type NoMoreConnectionsError struct {
	pos      Point
	otherPos Point
}

func (e *NoMoreConnectionsError) Error() string {
	return fmt.Sprintf("node at %v already has full connections and cannot connect to %v", e.pos, e.otherPos)
}

func (newPosEo *NewPossibleOutgrowth) String() string {
	return fmt.Sprintf("NP %v %d: %s, %d", newPosEo.pos, newPosEo.event.id, newPosEo.state.String(), newPosEo.distance)
}

func (eo *EventOutgrowth) String() string {
	return fmt.Sprintf("%v: %s, %d, %d", *(eo.pos), eo.state.String(), eo.distance, len(eo.fromList))
}

func (eo *EventOutgrowth) AddFrom(from *EventOutgrowth) {
	if eo.fromList == nil {
		eo.fromList = []*EventOutgrowth{from,}
	} else {
		eo.fromList = append(eo.fromList, from)
	}
}

func (eo *EventOutgrowth) HasFrom() bool {
	return eo.fromList != nil && len(eo.fromList) > 0
}

func (eo *EventOutgrowth) CameFromPoint(point Point) bool {
	if !eo.HasFrom() {
		return false
	}
	for _, from := range eo.fromList {
		if *(from.pos) == point {
			return true
		}
	}
	return false
}

func (eo *EventOutgrowth) IsRoot() bool {
	return !eo.HasFrom()
}

func (eo *EventOutgrowth) DistanceFromLatest(evt *Event) Distance {
	space := evt.space
	return Distance(space.currentTime - evt.created) - eo.distance
}

func (eo *EventOutgrowth) IsOld(evt *Event) bool {
	if eo.IsRoot() {
		// Root event always active
		return false
	}
	return eo.DistanceFromLatest(evt) >= evt.space.EventOutgrowthOldThreshold
}

func (eo *EventOutgrowth) IsActive(evt *Event) bool {
	if eo.IsRoot() {
		// Root event always active
		return true
	}
	return eo.DistanceFromLatest(evt) <= evt.space.EventOutgrowthThreshold
}
