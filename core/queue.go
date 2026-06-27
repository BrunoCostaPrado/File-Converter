package core

import (
	"encoding/json"
	"os"
	"sync"
)

type Queue struct {
	mu    sync.Mutex
	Items []QueueItem
}

func NewQueue() *Queue {
	return &Queue{Items: []QueueItem{}}
}

func (q *Queue) Add(items ...QueueItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Items = append(q.Items, items...)
}

func (q *Queue) NextPending() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.Items {
		if q.Items[i].Status == "pending" {
			q.Items[i].Status = "running"
			return &q.Items[i]
		}
	}
	return nil
}

func (q *Queue) UpdateStatus(index int, status string, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < len(q.Items) {
		q.Items[index].Status = status
		q.Items[index].Error = errMsg
	}
}

func (q *Queue) Save(path string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	data, err := json.MarshalIndent(q.Items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadQueue(path string) (*Queue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []QueueItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return &Queue{Items: items}, nil
}
