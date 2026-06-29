use eframe::egui;
use crate::preset_panel::PresetPanel;
use crate::queue_panel::QueuePanel;
use crate::settings::SettingsDialog;
use crate::source_panel::SourcePanel;

pub struct FileConverterApp {
    pub source_panel: SourcePanel,
    pub preset_panel: PresetPanel,
    pub queue_panel: QueuePanel,
    pub output_dir: String,
    pub ffmpeg_path: String,
    pub hwaccel: String,
    pub concurrent_jobs: usize,
    pub show_settings: bool,
}

impl FileConverterApp {
    pub fn new() -> Self {
        let output_dir = dirs::home_dir()
            .map(|p| p.join("Videos").to_string_lossy().to_string())
            .unwrap_or_else(|| ".".to_string());
        let ffmpeg_path = file_converter_core::ffmpeg::find_ffmpeg("");

        FileConverterApp {
            source_panel: SourcePanel::new(),
            preset_panel: PresetPanel::new(),
            queue_panel: QueuePanel::new(),
            output_dir,
            ffmpeg_path,
            hwaccel: String::new(),
            concurrent_jobs: 2,
            show_settings: false,
        }
    }

    // ponytail: stub - wire into WorkerPool when background processing is added
    pub fn start_queue(&mut self) {}
}

impl eframe::App for FileConverterApp {
    fn update(&mut self, ctx: &egui::Context, _frame: &mut eframe::Frame) {
        egui::SidePanel::left("source_panel")
            .resizable(true)
            .default_width(250.0)
            .show(ctx, |ui| {
                self.source_panel.ui(ui, &mut self.queue_panel);
            });

        egui::CentralPanel::default().show(ctx, |ui| {
            ui.horizontal(|ui| {
                ui.label("Output:");
                if ui.button("Browse").clicked() {
                    if let Some(dir) = rfd::FileDialog::new().pick_folder() {
                        self.output_dir = dir.to_string_lossy().to_string();
                    }
                }
                ui.label(&self.output_dir);
            });
            ui.separator();

            self.preset_panel.ui(ui);

            ui.separator();

            self.queue_panel.ui(ui);

            ui.horizontal(|ui| {
                if ui.button("▶ Start").clicked() {
                    self.start_queue();
                }
                if ui.button("⏹ Stop").clicked() {
                    self.queue_panel.stop();
                }
                if ui.button("Clear").clicked() {
                    self.queue_panel.clear();
                }
            });

            ui.separator();
            if ui.button("Settings").clicked() {
                self.show_settings = !self.show_settings;
            }
        });

        if self.show_settings {
            let mut open = true;
            egui::Window::new("Settings")
                .open(&mut open)
                .show(ctx, |ui| {
                    let mut dlg = SettingsDialog {
                        ffmpeg_path: &mut self.ffmpeg_path,
                        hwaccel: &mut self.hwaccel,
                        output_dir: &mut self.output_dir,
                        concurrent_jobs: &mut self.concurrent_jobs,
                    };
                    dlg.ui(ui);
                });
            if !open {
                self.show_settings = false;
            }
        }
    }
}
