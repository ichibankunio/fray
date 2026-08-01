package fray

import (
	"testing"

	"github.com/ichibankunio/fvec/vec3"
)

func TestBillboardSpatialIndexQueriesNearbyInstances(t *testing.T) {
	instances := []BillboardInstance{
		{Position: vec3.New(1, 1, 0)}, {Position: vec3.New(5, 1, 0)}, {Position: vec3.New(20, 20, 0)},
	}
	index := NewBillboardSpatialIndex(instances, 4)
	got := index.QueryInto(nil, 0, 0, 6)
	if index.Len() != 3 || len(got) != 2 {
		t.Fatalf("len=%d query=%d, want 3 and 2", index.Len(), len(got))
	}
}

func TestBillboardSpatialIndexReusesDestination(t *testing.T) {
	index := NewBillboardSpatialIndex([]BillboardInstance{{Position: vec3.New(1, 1, 0)}}, 4)
	buffer := make([]BillboardInstance, 0, 4)
	got := index.QueryInto(buffer, 1, 1, 1)
	if len(got) != 1 || cap(got) != cap(buffer) {
		t.Fatalf("query did not reuse destination: len=%d cap=%d", len(got), cap(got))
	}
}
