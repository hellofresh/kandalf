// The nearly all implementation of FSM is taken from
// https://github.com/siddontang/redis-failover project.
package cluster

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"

	"kandalf/logger"
)

const (
	cmdAdd = "add"
	cmdDel = "del"
	cmdSet = "set"
)

type fsm struct {
	sync.Mutex

	kvStore map[string]struct{}
}

type snapshot struct {
	nodes []string
}

func newFsm() *fsm {
	return &fsm{
		kvStore: make(map[string]struct{}),
	}
}

type action struct {
	Cmd     string   `json:"cmd"`
	Masters []string `json:"masters"`
}

func (f *fsm) Apply(l *raft.Log) interface{} {
	var a action

	if err := json.Unmarshal(l.Data, &a); err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to decode raft log")

		return err
	}

	f.handleAction(&a)

	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.Lock()
	defer f.Unlock()

	snap := new(snapshot)
	snap.nodes = make([]string, 0, len(f.kvStore))

	for node := range f.kvStore {
		snap.nodes = append(snap.nodes, node)
	}

	return snap, nil
}

func (f *fsm) Restore(snap io.ReadCloser) error {
	f.Lock()
	defer f.Unlock()
	defer snap.Close()

	d := json.NewDecoder(snap)
	var nodes []string

	if err := d.Decode(&nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		f.kvStore[node] = struct{}{}
	}

	return nil
}

func (f *fsm) handleAction(a *action) {
	switch a.Cmd {
	case cmdAdd:
		f.addNodes(a.Masters)
	case cmdDel:
		f.delNodes(a.Masters)
	case cmdSet:
		f.setNodes(a.Masters)
	}
}

func (f *fsm) addNodes(addrs []string) {
	f.Lock()
	defer f.Unlock()

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}

		f.kvStore[addr] = struct{}{}
	}
}

func (f *fsm) delNodes(addrs []string) {
	f.Lock()
	defer f.Unlock()

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}

		delete(f.kvStore, addr)
	}
}

func (f *fsm) setNodes(addrs []string) {
	f.Lock()
	defer f.Unlock()

	f.kvStore = make(map[string]struct{})

	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}

		f.kvStore[addr] = struct{}{}
	}
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	data, _ := json.Marshal(s.nodes)
	_, err := sink.Write(data)

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (s *snapshot) Release() {
}
