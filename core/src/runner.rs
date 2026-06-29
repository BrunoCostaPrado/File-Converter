use std::io::BufRead;
use std::process::{Command, Stdio};

use crate::convert::{build_ffmpeg_args, parse_progress_line};
use crate::types::Preset;
use crate::types::Progress;

pub fn run_transcode(
    ffmpeg_path: &str,
    input: &str,
    output: &str,
    preset: &Preset,
    on_progress: &mut dyn FnMut(Progress),
    bitrate: Option<&str>,
) -> Result<(), String> {
    let args = build_ffmpeg_args(input, output, preset, bitrate);
    let mut child = Command::new(ffmpeg_path)
        .args(&args)
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| format!("failed to spawn ffmpeg: {e}"))?;

    let stderr = child.stderr.take().ok_or_else(|| "no stderr".to_string())?;
    let reader = std::io::BufReader::new(stderr);
    let mut last_lines: Vec<String> = Vec::new();

    for line in reader.lines() {
        let line = line.map_err(|e| format!("failed to read stderr: {e}"))?;
        last_lines.push(line.clone());
        if last_lines.len() > 5 {
            last_lines.remove(0);
        }
        if parse_progress_line(&line).is_some() {
            on_progress(Progress {
                file: input.to_string(),
                percent: 0.0,
                status: "running".into(),
            });
        }
    }

    let status = child.wait().map_err(|e| format!("failed to wait on ffmpeg: {e}"))?;
    if status.success() {
        Ok(())
    } else {
        Err(last_lines.join("\n"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_run_transcode_invalid_ffmpeg() {
        let preset = Preset {
            name: "test".into(),
            container: "mp4".into(),
            video_codec: "h264".into(),
            audio_codec: "aac".into(),
            quality: 23,
            preset: "medium".into(),
            resolution: "1920x1080".into(),
            hwaccel: String::new(),
        };
        let result = run_transcode(
            "/nonexistent/ffmpeg",
            "in.mp4",
            "out.mp4",
            &preset,
            &mut |_| {},
            None,
        );
        assert!(result.is_err());
    }
}
