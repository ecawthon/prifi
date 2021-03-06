package time

import (
	"bytes"
	"encoding/gob"
)

//"time"

type LogEntry struct {
	Seq  uint64 // Consecutively-incrementing log entry sequence number
	Root HashId // Merkle root of values committed this time-step
	Time *int64 // Optional wall-clock time this entry was created
}

type SignedEntry struct {
	Ent []byte // Encoded LogEntry to which the signature applies
	Sig []byte // Digital signature on the LogEntry
}

type StampRequest struct {
	Val []byte // Hash-size value to timestamp
}
type StampReply struct {
	Sig []byte // Signature on the root
	// Prf Proof  // Merkle proof of value
}

// Request to obtain an old log-entry and, optionally,
// a cryptographic proof that it happened before a given newer entry.
// The server may be unable to process if Seq is beyond the retention window.
type EntryRequest struct {
	Seq uint64 // Sequence number of old entry requested
}
type EntryReply struct {
	Log SignedEntry // Signed log entry
}

// Request a cryptographic Merkle proof that log-entry Old happened before New.
// Produces a path to a Merkle tree node containing a hash of the node itself
// and the root of the history values committed within the node.
// The server may be unable to process if Old is beyond the retention window.
type ProofRequest struct {
	Old, New uint64 // Sequence number of old and new log records
}
type ProofReply struct {
	Prf Proof // Requested Merkle proof
}

// XXX not sure we need block requests?
type BlockRequest struct {
	Ids []HashId // Hash of block(s) requested
}

type BlockReply struct {
	Dat [][]byte // Content of block(s) requested
}

type ErrorReply struct {
	Msg string // Human-readable error message
}

type Message struct {
	ReqNo      uint64      // Request sequence number
	ErrorReply *ErrorReply // Generic error reply to any request

	StampRequest *StampRequest
	StampReply   *StampReply

	EntryRequest *EntryRequest
	EntryReply   *EntryReply

	ProofRequest *ProofRequest
	ProofReply   *ProofReply

	//BlockRequest *BlockRequest
	//BlockReply *BlockReply
}

type MessageType int

const (
	Error MessageType = iota
	StampRequestType
	StampReplyType
)

type TimeStampMessage struct {
	Type MessageType
	sreq *StampRequest
	srep *StampReply
}

func (tsm TimeStampMessage) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	var sub []byte
	var err error
	b.WriteByte(byte(tsm.Type))
	// marshal sub message based on its Type
	switch tsm.Type {
	case StampRequestType:
		sub, err = tsm.sreq.MarshalBinary()
	case StampReplyType:
		sub, err = tsm.srep.MarshalBinary()
	}
	if err == nil {
		b.Write(sub)
	}
	return b.Bytes(), err
}

func (sm *TimeStampMessage) UnmarshalBinary(data []byte) error {
	sm.Type = MessageType(data[0])
	msgBytes := data[1:]
	var err error
	switch sm.Type {
	case StampRequestType:
		sm.sreq = &StampRequest{}
		err = sm.sreq.UnmarshalBinary(msgBytes)
	case StampReplyType:
		sm.srep = &StampReply{}
		err = sm.srep.UnmarshalBinary(msgBytes)

	}
	return err
}

func (sreq StampRequest) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(sreq.Val)
	return b.Bytes(), err
}

func (sreq *StampRequest) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&sreq.Val)
	return err
}

func (srep StampReply) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(srep.Sig)
	return b.Bytes(), err
}

func (srep *StampReply) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&srep.Sig)
	return err
}
