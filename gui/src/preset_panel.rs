use eframe::egui;
use file_converter_core::preset::default_presets;

pub struct PresetPanel {
    pub selected: usize,
    pub preset_names: Vec<String>,
}

impl PresetPanel {
    pub fn new() -> Self {
        let names = default_presets().into_iter().map(|p| p.name).collect();
        PresetPanel {
            selected: 0,
            preset_names: names,
        }
    }

    pub fn ui(&mut self, ui: &mut egui::Ui) {
        ui.horizontal(|ui| {
            ui.label("Preset:");
            egui::ComboBox::from_id_source("preset_combo")
                .selected_text(&self.preset_names[self.selected])
                .show_ui(ui, |ui| {
                    for (i, name) in self.preset_names.iter().enumerate() {
                        ui.selectable_value(&mut self.selected, i, name);
                    }
                });
        });
    }
}
