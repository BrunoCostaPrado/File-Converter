use crate::types::Preset;
use regex::Regex;
use std::sync::LazyLock;

static VIDEO_ENCODERS: LazyLock<Vec<(&str, &str)>> = LazyLock::new(|| {
    vec![
        ("h264", "libx264"),
        ("h265", "libx265"),
        ("vp9", "libvpx-vp9"),
    ]
});

static HW_ENCODERS: LazyLock<Vec<(&str, Vec<(&str, &str)>)>> = LazyLock::new(|| {
    vec![
        ("nvenc", vec![("h264", "h264_nvenc"), ("h265", "hevc_nvenc")]),
        ("qsv", vec![("h264", "h264_qsv"), ("h265", "hevc_qsv")]),
        ("amd", vec![("h264", "h264_amf"), ("h265", "hevc_amf")]),
        ("videotoolbox", vec![("h264", "h264_videotoolbox"), ("h265", "hevc_videotoolbox")]),
    ]
});

// ponytail: each HW backend uses a different constant-quality flag
// videotoolbox -quality range is 1-100 (higher=better), not 0-51 like CRF.
static HW_QUALITY_FLAGS: LazyLock<Vec<(&str, &str)>> = LazyLock::new(|| {
    vec![
        ("nvenc", "-cq"),
        ("qsv", "-global_quality"),
        ("amd", "-quality"),
        ("videotoolbox", "-quality"),
    ]
});

pub fn encoder_name(hwaccel: &str, codec: &str) -> String {
    if codec == "copy" {
        return "copy".into();
    }
    if !hwaccel.is_empty() {
        for (hw_name, codecs) in HW_ENCODERS.iter() {
            if *hw_name == hwaccel {
                for (c, enc) in codecs.iter() {
                    if *c == codec {
                        return enc.to_string();
                    }
                }
            }
        }
    }
    for (c, enc) in VIDEO_ENCODERS.iter() {
        if *c == codec {
            return enc.to_string();
        }
    }
    codec.to_string()
}

pub fn build_ffmpeg_args(input: &str, output: &str, p: &Preset, bitrate: Option<&str>) -> Vec<String> {
    let mut args = vec!["-i".into(), input.to_string()];

    if p.video_codec != "copy" && !p.video_codec.is_empty() {
        let enc = encoder_name(&p.hwaccel, &p.video_codec);
        args.push("-c:v".into());
        args.push(enc);

        if p.hwaccel.is_empty() && !p.preset.is_empty() {
            args.push("-preset".into());
            args.push(p.preset.clone());
        }

        if let Some(br) = bitrate {
            args.push("-b:v".into());
            args.push(br.to_string());
        } else if let Some(flag) = HW_QUALITY_FLAGS.iter().find(|(k, _)| *k == p.hwaccel).map(|(_, v)| *v) {
            args.push(flag.into());
            args.push(p.quality.to_string());
        } else if p.hwaccel.is_empty() {
            args.push("-crf".into());
            args.push(p.quality.to_string());
        }
    } else if p.video_codec == "copy" {
        args.push("-c:v".into());
        args.push("copy".into());
    }

    if p.audio_codec != "copy" && !p.audio_codec.is_empty() {
        args.push("-c:a".into());
        args.push(p.audio_codec.clone());
    } else if p.audio_codec == "copy" {
        args.push("-c:a".into());
        args.push("copy".into());
    }

    if !p.resolution.is_empty() {
        let parts: Vec<&str> = p.resolution.split('x').collect();
        if parts.len() == 2 {
            args.push("-vf".into());
            args.push(format!("scale={}:{}", parts[0], parts[1]));
        }
    }

    args.push(output.to_string());
    args
}

static PROGRESS_REGEX: LazyLock<Regex> = LazyLock::new(|| {
    Regex::new(r"time=(\d+):(\d+):(\d+)\.(\d+)").unwrap()
});

pub fn parse_progress_line(line: &str) -> Option<f64> {
    let caps = PROGRESS_REGEX.captures(line)?;
    let hours: f64 = caps[1].parse().ok()?;
    let minutes: f64 = caps[2].parse().ok()?;
    let seconds: f64 = caps[3].parse().ok()?;
    let frac: f64 = caps[4].parse().ok()?;
    let divisor = 10.0_f64.powi(caps[4].len() as i32);
    Some(hours * 3600.0 + minutes * 60.0 + seconds + frac / divisor)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::types::Preset;

    #[test]
    fn test_encoder_name_cpu() {
        assert_eq!(encoder_name("", "h264"), "libx264");
    }

    #[test]
    fn test_encoder_name_nvenc() {
        assert_eq!(encoder_name("nvenc", "h264"), "h264_nvenc");
    }

    #[test]
    fn test_build_args_cpu() {
        let p = Preset {
            name: "test".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: "medium".into(),
            resolution: "1920x1080".into(),
            hwaccel: String::new(),
        };
        let args = build_ffmpeg_args("in.mp4", "out.mp4", &p, None);
        assert!(args.contains(&"-crf".to_string()));
        assert!(args.contains(&"23".to_string()));
    }

    #[test]
    fn test_build_args_nvenc_no_bitrate() {
        let p = Preset {
            name: "test".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: String::new(),
            resolution: "1920x1080".into(),
            hwaccel: "nvenc".into(),
        };
        let args = build_ffmpeg_args("in.mp4", "out.mp4", &p, None);
        assert!(args.contains(&"-cq".to_string()));
        assert!(args.contains(&"23".to_string()));
    }

    #[test]
    fn test_build_args_nvenc_with_bitrate() {
        let p = Preset {
            name: "test".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: String::new(),
            resolution: "1920x1080".into(),
            hwaccel: "nvenc".into(),
        };
        let args = build_ffmpeg_args("in.mp4", "out.mp4", &p, Some("2M"));
        assert!(args.contains(&"-b:v".to_string()));
        assert!(args.contains(&"2M".to_string()));
        assert!(!args.contains(&"-cq".to_string()));
    }

    #[test]
    fn test_build_args_copy() {
        let p = Preset {
            name: "test".into(),
            container: "mkv".into(),
            video_codec: "copy".into(),
            audio_codec: "copy".into(),
            quality: 0,
            preset: String::new(),
            resolution: String::new(),
            hwaccel: String::new(),
        };
        let args = build_ffmpeg_args("in.mp4", "out.mkv", &p, None);
        assert_eq!(args, vec!["-i", "in.mp4", "-c:v", "copy", "-c:a", "copy", "out.mkv"]);
    }

    #[test]
    fn test_parse_progress() {
        let line = "frame=123 fps=30 time=00:01:23.45 bitrate=1234.5kbits/s";
        let secs = parse_progress_line(line).unwrap();
        assert!((secs - 83.45).abs() < 0.01);
    }

    #[test]
    fn test_parse_progress_no_match() {
        assert!(parse_progress_line("ffmpeg version 4.4").is_none());
    }
}
