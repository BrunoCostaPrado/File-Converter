package core

import (
	"encoding/json"
	"testing"
)

func TestQueueItemJSONRoundTrip(t *testing.T) {
	item := QueueItem{InputPath: "a.mp4", Status: "pending"}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var decoded QueueItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.InputPath != "a.mp4" || decoded.Status != "pending" {
		t.Fatalf("round trip lost data: %+v", decoded)
	}
}
