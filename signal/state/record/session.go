package record

import (
	"bytes"

	"wa/signal/pb"

	"google.golang.org/protobuf/proto"
)

// archivedStatesMaxLength describes how many previous session
// states we should keep track of.
const archivedStatesMaxLength int = 40

// NewSessionFromBytes will return a Signal Session from the given
// bytes using the given serializer.
func NewSessionFromBytes(serialized []byte) (*Session, error) {
	p := &pb.RecordStructure{}
	e := proto.Unmarshal(serialized, p)
	if e != nil {
		return nil, e
	}

	return NewSessionFromStructure(p)
}

// NewSession creates a new session record and uses the given session and state
// serializers to convert the object into storeable bytes.
func NewSession() *Session {
	record := Session{
		sessionState:   NewState(),
		previousStates: []*State{},
		fresh:          true,
	}

	return &record
}

func NewSessionFromStructure(structure *pb.RecordStructure) (*Session, error) {

	previousStates := []*State{}
	for _, sess := range structure.GetPreviousSessions() {
		// proto.Unmarshal can return empty record
		if sess.SessionVersion == nil {
			continue
		}

		state, err := NewStateFromStructure(sess)
		if err != nil {
			return nil, err
		}
		previousStates = append(previousStates, state)
	}

	sessionState, err := NewStateFromStructure(structure.GetCurrentSession())
	if err != nil {
		return nil, err
	}

	session := &Session{
		previousStates: previousStates,
		sessionState:   sessionState,
		fresh:          false,
	}

	return session, nil
}

// NewSessionFromState creates a new session record from the given
// session state.
func NewSessionFromState(sessionState *State) *Session {
	record := Session{
		sessionState:   sessionState,
		previousStates: []*State{},
		fresh:          false,
	}

	return &record
}

// Session encapsulates the state of an ongoing session.
type Session struct {
	sessionState   *State
	previousStates []*State
	fresh          bool
}

// SetState sets the session record's current state to the given
// one.
func (r *Session) SetState(sessionState *State) {
	r.sessionState = sessionState
}

// IsFresh is used to determine if this is a brand new session
// or if a session record has already existed.
func (r *Session) IsFresh() bool {
	return r.fresh
}

// SessionState returns the session state object of the current
// session record.
func (r *Session) SessionState() *State {
	return r.sessionState
}

// PreviousSessionStates returns a list of all currently maintained
// "previous" session states.
func (r *Session) PreviousSessionStates() []*State {
	return r.previousStates
}

// HasSessionState will check this record to see if the sender's
// base key exists in the current and previous states.
func (r *Session) HasSessionState(version int, senderBaseKey []byte) bool {
	// Ensure the session state version is identical to this one.
	if r.sessionState.Version() == version && (bytes.Compare(senderBaseKey, r.sessionState.SenderBaseKey()) == 0) {
		return true
	}

	// Loop through all of our previous states and see if this
	// exists in our state.
	for i := range r.previousStates {
		if r.previousStates[i].Version() == version && bytes.Compare(senderBaseKey, r.previousStates[i].SenderBaseKey()) == 0 {
			return true
		}
	}

	return false
}

// ArchiveCurrentState moves the current session state into the list
// of "previous" session states, and replaces the current session state
// with a fresh reset instance.
func (r *Session) ArchiveCurrentState() {
	r.PromoteState(NewState())
}

// PromoteState takes the given session state and replaces it with the
// current state, pushing the previous current state to "previousStates".
func (r *Session) PromoteState(promotedState *State) {
	r.previousStates = r.prependStates(r.previousStates, r.sessionState)
	r.sessionState = promotedState

	// Remove the last state if it has reached our maximum length
	if len(r.previousStates) > archivedStatesMaxLength {
		r.previousStates = r.removeLastState(r.previousStates)
	}
}

// Serialize will return the session as serialized bytes so it can be
// persistently stored.
func (r *Session) Serialize() ([]byte, error) {
	return proto.Marshal(r.structure())
}

// prependStates takes an array/slice of states and prepends it with
// the given session state.
func (r *Session) prependStates(states []*State, sessionState *State) []*State {
	return append([]*State{sessionState}, states...)
}

// removeLastState takes an array/slice of states and removes the
// last element from it.
func (r *Session) removeLastState(states []*State) []*State {
	return states[:len(states)-1]
}

// structure will return a simple serializable session structure
// from the given structure. This is used for serialization to persistently
// store a session record.
func (r *Session) structure() *pb.RecordStructure {
	previousStates := make([]*pb.SessionStructure, len(r.previousStates))
	for i := range r.previousStates {
		previousStates[i] = r.previousStates[i].structure()
	}
	return &pb.RecordStructure{
		CurrentSession:   r.sessionState.structure(),
		PreviousSessions: previousStates,
	}
}
