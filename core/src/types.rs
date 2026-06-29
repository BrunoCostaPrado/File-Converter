use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Preset {
    pub name: String,
    pub container: String,
    pub video_codec: String,
    pub audio_codec: String,
    pub quality: i32,
    pub preset: String,
    pub resolution: String,
    pub hwaccel: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QueueItem {
    pub input_path: String,
    pub output_path: String,
    pub preset_name: String,
    pub status: String,
    pub progress: f64,
    pub error: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Progress {
    pub file: String,
    pub percent: f64,
    pub status: String,
}
