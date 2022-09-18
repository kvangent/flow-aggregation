package main

import (
	"context"
	"encoding/json"
	"testing"
)

// TestFlow verifies that the Flow struct serializes correctly.
func TestFlow(t *testing.T) {
	tcs := []struct {
		input string
		want  Flow
	}{
		{
			`{
				"src_app":"foo",
				"dest_app":"bar",
				"vpc_id":"vpc-0",
				"bytes_tx":100,
				"bytes_rx":500,
				"hour":1
			 }`,
			NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
		},
	}

	for _, tc := range tcs {
		got := Flow{}
		err := json.Unmarshal([]byte(tc.input), &got)
		if err != nil {
			t.Fatalf("unable to deseriaize: %v", err)
		}
		if tc.want != got {
			t.Fatalf("unexpected results: got %v, want %v", got, tc.want)
		}
	}
}

// TestMemoryContrller tests the MemoryController's functionality.
func TestMemoryController(t *testing.T) {
	tcs := []struct {
		desc  string
		input []Flow
		want  []Flow
	}{
		{
			"test two flows that should aggregate together",
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
			},
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 200, 1000),
			},
		},
		{
			"test two flows with different vpcs",
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-1", "foo", "bar", 1, 100, 500),
			},
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-1", "foo", "bar", 1, 100, 500),
			},
		},
		{
			"test two flows with different source apps",
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "buzz", "bar", 1, 100, 500),
			},
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "buzz", "bar", 1, 100, 500),
			},
		},
		{
			"test two flows with different destination apps",
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "foo", "buzz", 1, 100, 500),
			},
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "foo", "buzz", 1, 100, 500),
			},
		},
		{
			"test two flows with different hours",
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "foo", "bar", 2, 100, 500),
			},
			[]Flow{
				NewFlow("vpc-0", "foo", "bar", 1, 100, 500),
				NewFlow("vpc-0", "foo", "bar", 2, 100, 500),
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			c := NewMemoryController()
			if err := c.FlowAggregate(ctx, tc.want); err != nil {
				t.Fatalf("error during Aggregate: %v", err)
			}

			got, err := c.FlowReadAll(ctx)
			if err != nil {
				t.Fatalf("error during ReadAll: %v", err)
			}
			if !equalsUnordered(tc.want, got) {
				t.Fatalf("ReadAll returned unexpected result: got %v, want %v", got, tc.want)

			}
		})
	}
}

// equalsUnordered returns true if the Flow in a slice are equal, regardless of order.
func equalsUnordered(a []Flow, b []Flow) bool {
	if len(a) != len(b) {
		return false
	}
	// Add all the elements of a into a map, where flow -> count
	e := make(map[Flow]int)
	for _, v := range a {
		e[v] += 1
	}
	// remove all of the elements of b
	for _, v := range b {
		e[v] -= 1
		if e[v] == 0 {
			delete(e, v)
		}
	}
	// they are equal if the map is empty
	return len(e) == 0
}
