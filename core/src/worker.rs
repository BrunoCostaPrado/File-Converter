use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::mpsc;
use std::sync::Arc;
use std::thread;

use crate::preset::find_preset;
use crate::runner::run_transcode;
use crate::types::{Progress, QueueItem};

pub struct WorkerPool {
    canceled: Arc<AtomicBool>,
}

impl WorkerPool {
    pub fn new() -> Self {
        WorkerPool {
            canceled: Arc::new(AtomicBool::new(false)),
        }
    }

    pub fn start(
        &self,
        ffmpeg_path: &str,
        global_hwaccel: &str,
        items: Vec<QueueItem>,
        concurrent: usize,
    ) -> Vec<QueueItem> {
        let n = concurrent.max(1);
        let len = items.len();
        self.canceled.store(false, Ordering::Relaxed);

        let (tx, rx) = mpsc::channel::<(usize, QueueItem)>();
        let items = Arc::new(items);
        let mut handles = Vec::with_capacity(n);

        for thread_id in 0..n {
            let tx = tx.clone();
            let items = Arc::clone(&items);
            let canceled = Arc::clone(&self.canceled);
            let ffmpeg_path = ffmpeg_path.to_string();
            let global_hwaccel = global_hwaccel.to_string();

            handles.push(thread::spawn(move || {
                for i in (thread_id..len).step_by(n) {
                    if canceled.load(Ordering::Relaxed) {
                        break;
                    }

                    let item = &items[i];
                    let mut preset = find_preset(&item.preset_name);
                    if preset.hwaccel.is_empty() && !global_hwaccel.is_empty() {
                        preset.hwaccel = global_hwaccel.clone();
                    }

                    let result = run_transcode(
                        &ffmpeg_path,
                        &item.input_path,
                        &item.output_path,
                        &preset,
                        &mut |progress: Progress| {
                            let mut qi = item.clone();
                            qi.status = "running".into();
                            qi.progress = progress.percent;
                            let _ = tx.send((i, qi));
                        },
                        None,
                    );

                    let mut final_item = item.clone();
                    match result {
                        Ok(()) => {
                            final_item.status = "completed".into();
                            final_item.progress = 100.0;
                        }
                        Err(e) => {
                            final_item.status = "failed".into();
                            final_item.error = e;
                        }
                    }
                    let _ = tx.send((i, final_item));
                }
            }));
        }

        drop(tx);

        let mut results: Vec<Option<QueueItem>> = (0..len).map(|_| None).collect();
        for (i, item) in rx {
            results[i] = Some(item);
        }

        for handle in handles {
            let _ = handle.join();
        }

        results
            .into_iter()
            .enumerate()
            .map(|(i, r)| {
                r.unwrap_or_else(|| {
                    let mut item = (*items)[i].clone();
                    item.status = "stopped".into();
                    item
                })
            })
            .collect()
    }

    pub fn stop(&self) {
        self.canceled.store(true, Ordering::Relaxed);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_worker_pool_empty_items() {
        let pool = WorkerPool::new();
        let results = pool.start("ffmpeg", "", vec![], 4);
        assert!(results.is_empty());
    }

    #[test]
    fn test_worker_pool_invalid_ffmpeg() {
        let pool = WorkerPool::new();
        let items = vec![QueueItem {
            input_path: "in.mp4".into(),
            output_path: "out.mp4".into(),
            preset_name: "Fast 1080p30".into(),
            status: "pending".into(),
            progress: 0.0,
            error: String::new(),
        }];
        let results = pool.start("/nonexistent/ffmpeg", "", items, 1);
        assert_eq!(results.len(), 1);
        assert_eq!(results[0].status, "failed");
    }
}
