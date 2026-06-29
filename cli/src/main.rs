use std::io::{self, Write};
use std::path::Path;

use clap::Parser;
use file_converter_core::ffmpeg::{find_ffmpeg, probe_bitrate};
use file_converter_core::preset::find_preset;
use file_converter_core::runner::run_transcode;
use file_converter_core::types::{Progress, QueueItem};
use file_converter_core::worker::WorkerPool;

#[derive(Parser)]
#[command(name = "file_converter", version)]
struct Cli {
    #[arg(short, long, default_value = "Fast 1080p30")]
    preset: String,

    #[arg(short, long, default_value = "./output")]
    output: String,

    #[arg(long, default_value = "")]
    ffmpeg_path: String,

    #[arg(long, default_value = "")]
    hwaccel: String,

    #[arg(long, default_value = "")]
    bitrate: String,

    #[arg(long, default_value_t = false)]
    keep_format: bool,

    #[arg(long, default_value_t = false)]
    queue: bool,

    #[arg(short, long, default_value_t = 2)]
    concurrent: usize,

    #[arg(trailing_var_arg = true)]
    inputs: Vec<String>,
}

fn main() {
    let cli = Cli::parse();

    let ffmpeg = find_ffmpeg(&cli.ffmpeg_path);
    if ffmpeg.is_empty() {
        eprintln!("ffmpeg not found");
        std::process::exit(1);
    }

    if cli.queue {
        let mut input = String::new();
        io::stdin().read_line(&mut input).ok();
        let items: Vec<QueueItem> = serde_json::from_str(&input).unwrap_or_default();
        let pool = WorkerPool::new();
        let results = pool.start(&ffmpeg, &cli.hwaccel, items, cli.concurrent);
        println!("{}", serde_json::to_string_pretty(&results).unwrap());
        return;
    }

    if cli.inputs.is_empty() {
        eprintln!("No input files. Use --help for usage.");
        std::process::exit(1);
    }

    let default_output = Path::new("./output");

    for input in &cli.inputs {
        let mut preset = find_preset(&cli.preset);
        if !cli.hwaccel.is_empty() {
            preset.hwaccel = cli.hwaccel.clone();
        }

        let input_path = Path::new(input);
        let stem = input_path
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("output");

        let output = if cli.keep_format {
            let ext = input_path
                .extension()
                .and_then(|s| s.to_str())
                .unwrap_or(&preset.container);
            if Path::new(&cli.output) == default_output {
                format!("{}.{}", stem, ext)
            } else {
                format!("{}/{}.{}", cli.output, stem, ext)
            }
        } else if Path::new(&cli.output) == default_output {
            format!("{}.{}", stem, preset.container)
        } else {
            format!("{}/{}.{}", cli.output, stem, preset.container)
        };

        let bitrate = if !preset.hwaccel.is_empty() && cli.bitrate.is_empty() {
            probe_bitrate(&ffmpeg, input)
        } else if !cli.bitrate.is_empty() {
            Some(cli.bitrate.clone())
        } else {
            None
        };

        match run_transcode(
            &ffmpeg,
            input,
            &output,
            &preset,
            &mut |p: Progress| {
                print!("\r  {}: {:.0}%", p.file, p.percent);
                io::stdout().flush().ok();
            },
            bitrate.as_deref(),
        ) {
            Ok(()) => {
                println!("\n  done: {}", output);
            }
            Err(e) => {
                eprintln!("error: {}", e);
                std::process::exit(1);
            }
        }
    }
}
