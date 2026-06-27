package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var videoEncoders = map[string]string{
	"h264": "libx264",
	"h265": "libx265",
	"vp9":  "libvpx-vp9",
	"copy": "copy",
}

var hwEncoders = map[string]map[string]string{
	"nvenc":       {"h264": "h264_nvenc", "h265": "hevc_nvenc"},
	"qsv":         {"h264": "h264_qsv", "h265": "hevc_qsv"},
	"amd":         {"h264": "h264_amf", "h265": "hevc_amf"},
	"videotoolbox": {"h264": "h264_videotoolbox", "h265": "hevc_videotoolbox"},
}

func BuildFfmpegArgs(input, output string, p Preset, bitrate string) []string {
	args := []string{"-i", input}

	if p.VideoCodec != "" && p.VideoCodec != "copy" {
		enc := EncoderName(p.HWAccel, p.VideoCodec)
		args = append(args, "-c:v", enc)
		if p.HWAccel == "" && p.Preset != "" {
			args = append(args, "-preset", p.Preset)
		}
		if bitrate != "" {
			args = append(args, "-b:v", bitrate)
		} else if p.HWAccel == "" {
			args = append(args, "-crf", strconv.Itoa(p.Quality))
		}
	} else if p.VideoCodec == "copy" {
		args = append(args, "-c:v", "copy")
	}

	if p.AudioCodec != "" && p.AudioCodec != "copy" {
		args = append(args, "-c:a", p.AudioCodec)
	} else if p.AudioCodec == "copy" {
		args = append(args, "-c:a", "copy")
	}

	if p.Resolution != "" {
		args = append(args, "-vf", fmt.Sprintf("scale=%s", strings.Replace(p.Resolution, "x", ":", 1)))
	}

	args = append(args, output)
	return args
}

func EncoderName(hwAccel, codec string) string {
	if codec == "copy" {
		return "copy"
	}
	if hwAccel != "" {
		if m, ok := hwEncoders[hwAccel]; ok {
			if v, ok := m[codec]; ok {
				return v
			}
		}
	}
	if v, ok := videoEncoders[codec]; ok {
		return v
	}
	return codec
}

var progressRegex = regexp.MustCompile(`time=(\d+):(\d+):(\d+)\.(\d+)`)

func ParseProgressLine(line string) (float64, bool) {
	matches := progressRegex.FindStringSubmatch(line)
	if len(matches) < 5 {
		return 0, false
	}
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	centisecs, _ := strconv.Atoi(matches[4])
	total := float64(hours*3600+minutes*60+seconds) + float64(centisecs)/100.0
	return total, true
}
