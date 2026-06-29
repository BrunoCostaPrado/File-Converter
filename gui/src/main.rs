mod app;
mod source_panel;
mod preset_panel;
mod queue_panel;
mod settings;

fn main() {
    let native_options = eframe::NativeOptions {
        viewport: egui::ViewportBuilder::default()
            .with_inner_size([900.0, 600.0]),
        ..Default::default()
    };
    eframe::run_native(
        "File Converter",
        native_options,
        Box::new(|_cc| Box::new(app::FileConverterApp::new())),
    )
    .expect("eframe failed");
}
