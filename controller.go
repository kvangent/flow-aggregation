package main

import (
	"context"
	"sync"
)

// Flow represnts the Flow from an app to another app.
type Flow struct {
	flowUnique    `json:",inline"`
	flowCumlative `json:",inline"`
}

// NewFlow intializes a new Flow struct.
func NewFlow(vpcID, srcApp, destApp string, hour uint64, bytesTx, bytesRx uint64) Flow {
	return Flow{
		flowUnique{
			VpcID: vpcID, SrcApp: srcApp, DestApp: destApp, Hour: hour,
		},
		flowCumlative{
			BytesTx: bytesTx, BytesRx: 500,
		},
	}
}

// flowUnique represents a subset of fields in a Flow struct used for hashing.
type flowUnique struct {
	VpcID   string `json:"vpc_id"`
	SrcApp  string `json:"src_app"`
	DestApp string `json:"dest_app"`
	Hour    uint64 `json:"hour"`
}

type flowCumlative struct {
	BytesTx uint64 `json:"bytes_tx"`
	BytesRx uint64 `json:"bytes_rx"`
}

// Controller provides a common abstration for interacting with a datasource.
type Controller interface {
	FlowAggregate(context.Context, []Flow) error
	FlowReadAll(context.Context) ([]Flow, error)
}

// MemoryController is a an implementation of Controller that only persists in
// memory. It's theadsafe, but not very efficient or persistent. Intended for
// use with testing.
type MemoryController struct {
	// This could likely be optimized by using a RWMutex for the map,
	// and a seperate mutex for mutation of the Flow itself.
	mu   sync.Mutex
	data map[flowUnique]Flow
}

// verify interface
var _ Controller = (*MemoryController)(nil)

// NewMemoryController returns an initalized MemoryController.
func NewMemoryController() *MemoryController {
	return &MemoryController{data: make(map[flowUnique]Flow)}
}

// FlowAggregate adds a flow to be aggregated with the others.
func (m *MemoryController) FlowAggregate(ctx context.Context, fs []Flow) error {
	// TODO: lock doesn't respect ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range fs {
		v := m.data[f.flowUnique]
		// add the existing values (v) to the new values to present flowUnique
		f.BytesTx += v.BytesTx
		f.BytesRx += v.BytesRx
		m.data[f.flowUnique] = f
	}
	return nil
}

func (m *MemoryController) FlowReadAll(ctx context.Context) ([]Flow, error) {
	// TODO: lock doesn't respect ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Flow, 0, len(m.data))
	for _, f := range m.data {
		out = append(out, f)
	}
	return out, nil
}
