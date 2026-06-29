use crate::types::QueueItem;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Queue {
    pub items: Vec<QueueItem>,
}

impl Queue {
    pub fn new() -> Self {
        Queue { items: Vec::new() }
    }

    pub fn add(&mut self, items: Vec<QueueItem>) {
        self.items.extend(items);
    }

    pub fn next_pending(&mut self) -> Option<&mut QueueItem> {
        self.items.iter_mut().find(|item| item.status == "pending")
    }

    pub fn save(&self, path: &str) -> Result<(), String> {
        if let Some(parent) = std::path::Path::new(path).parent() {
            std::fs::create_dir_all(parent)
                .map_err(|e| format!("Failed to create queue dir: {}", e))?;
        }
        let json = serde_json::to_string_pretty(&self.items)
            .map_err(|e| format!("Failed to serialize queue: {}", e))?;
        std::fs::write(path, json)
            .map_err(|e| format!("Failed to write queue: {}", e))?;
        Ok(())
    }

    pub fn load(path: &str) -> Result<Self, String> {
        let content =
            std::fs::read_to_string(path).map_err(|e| format!("Failed to read queue: {}", e))?;
        let items: Vec<QueueItem> =
            serde_json::from_str(&content).map_err(|e| format!("Failed to parse queue: {}", e))?;
        Ok(Queue { items })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::QueueItem;

    fn make_item(name: &str, status: &str) -> QueueItem {
        QueueItem {
            input_path: format!("{}.mp4", name),
            output_path: format!("{}_out.mp4", name),
            preset_name: "Fast 1080p30".to_string(),
            status: status.to_string(),
            progress: 0.0,
            error: String::new(),
        }
    }

    #[test]
    fn test_queue_add_and_next_pending() {
        let mut q = Queue::new();
        assert!(q.next_pending().is_none());

        q.add(vec![
            make_item("a", "done"),
            make_item("b", "pending"),
            make_item("c", "pending"),
        ]);
        let next = q.next_pending();
        assert!(next.is_some());
        assert_eq!(next.unwrap().input_path, "b.mp4");
    }

    #[test]
    fn test_queue_save_load() {
        let dir = std::env::temp_dir().join("fc_test_queue");
        let path = dir.join("queue.json");
        let path_str = path.to_str().unwrap().to_string();

        let mut q = Queue::new();
        q.add(vec![make_item("test", "pending")]);
        q.save(&path_str).unwrap();

        let loaded = Queue::load(&path_str).unwrap();
        assert_eq!(loaded.items.len(), 1);
        assert_eq!(loaded.items[0].input_path, "test.mp4");
        assert_eq!(loaded.items[0].status, "pending");

        let _ = std::fs::remove_dir_all(&dir);
    }
}
