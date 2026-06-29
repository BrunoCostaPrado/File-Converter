use serde::{Deserialize, Serialize};
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Settings {
    pub ffmpeg_path: String,
    pub hwaccel: String,
    pub output_dir: String,
    pub default_preset: String,
    pub concurrent_jobs: i32,
}

impl Default for Settings {
    fn default() -> Self {
        let output_dir = dirs::home_dir()
            .map(|p| p.join("Videos").to_string_lossy().to_string())
            .unwrap_or_else(|| "~/Videos".to_string());
        Settings {
            ffmpeg_path: String::new(),
            hwaccel: String::new(),
            output_dir,
            default_preset: "Fast 1080p30".to_string(),
            concurrent_jobs: 2,
        }
    }
}

impl Settings {
    pub fn settings_path() -> PathBuf {
        // dirs::config_dir uses the correct path on all platforms:
        //   Windows: %APPDATA%
        //   macOS:   ~/Library/Application Support
        //   Linux:   $XDG_CONFIG_HOME or ~/.config
        dirs::config_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("file_converter")
            .join("settings.json")
    }

    pub fn save(&self, path: &str) -> Result<(), String> {
        if let Some(parent) = std::path::Path::new(path).parent() {
            std::fs::create_dir_all(parent)
                .map_err(|e| format!("Failed to create config dir: {}", e))?;
        }
        let json = serde_json::to_string_pretty(self)
            .map_err(|e| format!("Failed to serialize settings: {}", e))?;
        std::fs::write(path, json)
            .map_err(|e| format!("Failed to write settings: {}", e))?;
        Ok(())
    }

    pub fn load(path: &str) -> Self {
        std::fs::read_to_string(path)
            .ok()
            .and_then(|content| serde_json::from_str(&content).ok())
            .unwrap_or_default()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_settings_default() {
        let s = Settings::default();
        assert_eq!(s.ffmpeg_path, "");
        assert_eq!(s.hwaccel, "");
        assert!(s.output_dir.contains("Videos"));
        assert_eq!(s.default_preset, "Fast 1080p30");
        assert_eq!(s.concurrent_jobs, 2);
    }

    #[test]
    fn test_settings_save_load() {
        let dir = std::env::temp_dir().join("fc_test_settings");
        let path = dir.join("settings.json");
        let path_str = path.to_str().unwrap().to_string();

        let s = Settings {
            ffmpeg_path: "/usr/bin/ffmpeg".to_string(),
            hwaccel: "cuda".to_string(),
            output_dir: "/tmp/vids".to_string(),
            default_preset: "Slow 4K".to_string(),
            concurrent_jobs: 4,
        };
        s.save(&path_str).unwrap();

        let loaded = Settings::load(&path_str);
        assert_eq!(loaded.ffmpeg_path, "/usr/bin/ffmpeg");
        assert_eq!(loaded.hwaccel, "cuda");
        assert_eq!(loaded.output_dir, "/tmp/vids");
        assert_eq!(loaded.default_preset, "Slow 4K");
        assert_eq!(loaded.concurrent_jobs, 4);

        let _ = std::fs::remove_dir_all(&dir);
    }
}
