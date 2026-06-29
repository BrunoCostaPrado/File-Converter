use crate::queue_panel::QueuePanel;
use eframe::egui;

pub struct SourcePanel {
    pub files: Vec<String>,
}

impl SourcePanel {
    pub fn new() -> Self {
        Self { files: vec![] }
    }

    pub fn ui(&mut self, ui: &mut egui::Ui, queue: &mut QueuePanel) {
        ui.heading("Files");
        if ui.button("+ Add Files").clicked() {
            if let Some(files) = rfd::FileDialog::new().pick_files() {
                for f in &files {
                    let path = f.to_string_lossy().to_string();
                    self.files.push(path.clone());
                    queue.add_file(path);
                }
            }
        }
        let mut remove_idx: Vec<usize> = vec![];
        egui::ScrollArea::vertical().show(ui, |ui| {
            for (i, file) in self.files.iter().enumerate() {
                ui.horizontal(|ui| {
                    ui.label(file);
                    if ui.button("✕").clicked() {
                        remove_idx.push(i);
                    }
                });
            }
        });
        for i in remove_idx.into_iter().rev() {
            self.files.remove(i);
        }
    }
}
