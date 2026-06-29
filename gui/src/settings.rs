use eframe::egui;

pub struct SettingsDialog<'a> {
    pub ffmpeg_path: &'a mut String,
    pub hwaccel: &'a mut String,
    pub output_dir: &'a mut String,
    pub concurrent_jobs: &'a mut usize,
}

impl<'a> SettingsDialog<'a> {
    pub fn ui(&mut self, ui: &mut egui::Ui) {
        ui.label("FFmpeg Path:");
        ui.text_edit_singleline(self.ffmpeg_path);

        ui.label("HW Accel:");
        egui::ComboBox::from_id_source("hwaccel_combo")
            .selected_text(if self.hwaccel.is_empty() {
                "None (CPU)"
            } else {
                &self.hwaccel
            })
            .show_ui(ui, |ui| {
                let opts = ["", "nvenc", "qsv", "amd", "videotoolbox"];
                let mut selected = self.hwaccel.clone();
                for opt in &opts {
                    let label = if opt.is_empty() { "None (CPU)" } else { opt };
                    ui.selectable_value(&mut selected, opt.to_string(), label);
                }
                *self.hwaccel = selected;
            });

        ui.label("Output Dir:");
        ui.text_edit_singleline(self.output_dir);

        ui.label("Concurrent Jobs:");
        ui.add(egui::Slider::new(self.concurrent_jobs, 1..=8));
    }
}
