use eframe::egui;
use file_converter_core::types::{Progress, QueueItem};

pub struct QueuePanel {
    pub items: Vec<QueueItem>,
}

impl QueuePanel {
    pub fn new() -> Self {
        Self { items: vec![] }
    }

    pub fn add_file(&mut self, path: String) {
        let input = std::path::Path::new(&path);
        let stem = input
            .file_stem()
            .map(|s| s.to_string_lossy())
            .unwrap_or_default()
            .to_string();
        self.items.push(QueueItem {
            input_path: path,
            output_path: format!("output/{}.mp4", stem),
            preset_name: "Fast 1080p30".into(),
            status: "pending".into(),
            progress: 0.0,
            error: String::new(),
        });
    }

    pub fn update_progress(&mut self, prog: Progress) {
        if let Some(item) = self
            .items
            .iter_mut()
            .find(|i| i.input_path == prog.file)
        {
            item.progress = prog.percent;
            item.status = prog.status;
        }
    }

    // ponytail: stub - WorkerPool::stop will wire here
    pub fn stop(&mut self) {}

    pub fn clear(&mut self) {
        self.items.clear();
    }

    pub fn ui(&mut self, ui: &mut egui::Ui) {
        ui.heading("Queue");
        egui::ScrollArea::vertical().show(ui, |ui| {
            for item in &self.items {
                ui.group(|ui| {
                    ui.label(&item.input_path);
                    ui.label(format!("Status: {}", item.status));
                    if item.progress > 0.0 && item.progress < 100.0 {
                        ui.add(
                            egui::ProgressBar::new(item.progress as f32 / 100.0)
                                .text(format!("{:.0}%", item.progress)),
                        );
                    }
                });
            }
        });
    }
}
