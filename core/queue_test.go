package core

import (
	"path/filepath"
	"testing"
)

func TestQueueSaveLoad(t *testing.T) {
	q := &Queue{Items: []QueueItem{{InputPath: "a.mp4", Status: "pending"}}}
	path := filepath.Join(t.TempDir(), "queue.json")
	if err := q.Save(path); err != nil {
		t.Fatal(err)
	}
	q2, err := LoadQueue(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(q2.Items) != 1 || q2.Items[0].InputPath != "a.mp4" {
		t.Fatalf("round trip failed: %+v", q2.Items)
	}
}

func TestQueueNextItem(t *testing.T) {
	q := &Queue{
		Items: []QueueItem{
			{InputPath: "a.mp4", Status: "done"},
			{InputPath: "b.mp4", Status: "pending"},
			{InputPath: "c.mp4", Status: "pending"},
		},
	}
	item := q.NextPending()
	if item == nil || item.InputPath != "b.mp4" {
		t.Fatalf("expected b.mp4, got %v", item)
	}
}
