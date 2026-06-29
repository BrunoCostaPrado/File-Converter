use crate::types::Preset;

pub fn default_presets() -> Vec<Preset> {
    vec![
        Preset {
            name: "Fast 1080p30".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 22,
            preset: "fast".into(),
            resolution: "1920x1080".into(),
            hwaccel: String::new(),
        },
        Preset {
            name: "H.265 1080p".into(),
            container: "mkv".into(),
            video_codec: "h265".into(),
            audio_codec: "aac".into(),
            quality: 24,
            preset: "medium".into(),
            resolution: "1920x1080".into(),
            hwaccel: String::new(),
        },
        Preset {
            name: "Super HQ 1080p".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 18,
            preset: "slow".into(),
            resolution: "1920x1080".into(),
            hwaccel: String::new(),
        },
        Preset {
            name: "Very Fast 720p".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 22,
            preset: "veryfast".into(),
            resolution: "1280x720".into(),
            hwaccel: String::new(),
        },
        Preset {
            name: "Lossless".into(),
            container: "mkv".into(),
            video_codec: "h264".into(),
            audio_codec: "copy".into(),
            quality: 0,
            preset: "slow".into(),
            resolution: String::new(),
            hwaccel: String::new(),
        },
        Preset {
            name: "Audio Only".into(),
            container: "m4a".into(),
            video_codec: "copy".into(),
            audio_codec: "aac".into(),
            quality: 192,
            preset: String::new(),
            resolution: String::new(),
            hwaccel: String::new(),
        },
        Preset {
            name: "Copy Stream".into(),
            container: "mkv".into(),
            video_codec: "copy".into(),
            audio_codec: "copy".into(),
            quality: 0,
            preset: String::new(),
            resolution: String::new(),
            hwaccel: String::new(),
        },
        Preset {
            name: "NVENC 1080p".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: String::new(),
            resolution: "1920x1080".into(),
            hwaccel: "nvenc".into(),
        },
        Preset {
            name: "NVENC H.265 1080p".into(),
            container: "mkv".into(),
            video_codec: "h265".into(),
            audio_codec: "aac".into(),
            quality: 24,
            preset: String::new(),
            resolution: "1920x1080".into(),
            hwaccel: "nvenc".into(),
        },
        Preset {
            name: "AMD 1080p".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: String::new(),
            resolution: "1920x1080".into(),
            hwaccel: "amd".into(),
        },
    ]
}

pub fn find_preset(name: &str) -> Preset {
    let presets = default_presets();
    presets.into_iter().find(|p| p.name == name).unwrap_or_else(|| {
        let mut p = default_presets();
        p.remove(0)
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_presets_not_empty() {
        assert!(!default_presets().is_empty());
    }

    #[test]
    fn test_find_preset_known() {
        let p = find_preset("Fast 1080p30");
        assert_eq!(p.container, "mp4");
    }

    #[test]
    fn test_find_preset_unknown_falls_to_first() {
        let p = find_preset("nonexistent");
        assert_eq!(p.name, "Fast 1080p30");
    }
}
