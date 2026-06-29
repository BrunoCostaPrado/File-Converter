use std::path::Path;
use std::process::Command;

pub fn find_ffmpeg(user_path: &str) -> String {
    if !user_path.is_empty() {
        if Path::new(user_path).exists() {
            return user_path.to_string();
        }
        return String::new();
    }
    if let Ok(p) = which::which("ffmpeg") {
        return p.to_string_lossy().to_string();
    }
    String::new()
}

pub fn probe_bitrate(ffmpeg_path: &str, input: &str) -> Option<String> {
    let dir = Path::new(ffmpeg_path).parent()?;
    let stem = Path::new(ffmpeg_path).file_stem()?.to_str()?;
    let probe_name = stem.replace("ffmpeg", "ffprobe");
    let probe_path = dir.join(&probe_name);
    let probe = if probe_path.exists() {
        probe_path.to_string_lossy().to_string()
    } else if let Ok(p) = which::which("ffprobe") {
        p.to_string_lossy().to_string()
    } else {
        return None;
    };

    let out = Command::new(probe)
        .args([
            "-v", "error",
            "-select_streams", "v:0",
            "-show_entries", "stream=bit_rate",
            "-of", "default=noprint_wrappers=1:nokey=1",
            input,
        ])
        .output()
        .ok()?;

    let s = String::from_utf8_lossy(&out.stdout).trim().to_string();
    if s.is_empty() { None } else { Some(s) }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_find_ffmpeg_user_path_nonexistent() {
        assert_eq!(find_ffmpeg("/nonexistent/ffmpeg"), "");
    }

    #[test]
    fn test_find_ffmpeg_empty() {
        let _result = find_ffmpeg("");
        // may or may not find ffmpeg on PATH — just verify no crash
    }
}
