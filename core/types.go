package core

type Preset struct {
	Name       string `json:"name"`
	Container  string `json:"container"`   // mp4, mkv, webm
	VideoCodec string `json:"video_codec"` // h264, h265, vp9, copy
	AudioCodec string `json:"audio_codec"` // aac, opus, mp3, copy
	Quality    int    `json:"quality"`     // CRF 0-51
	Preset     string `json:"preset"`      // ultrafast..veryslow
	Resolution string `json:"resolution"`  // "" or "1920x1080"
	HWAccel    string `json:"hwaccel"`     // "", nvenc, qsv, amd, videotoolbox
}

type QueueItem struct {
	InputPath  string  `json:"input_path"`
	OutputPath string  `json:"output_path"`
	PresetName string  `json:"preset_name"`
	Status     string  `json:"status"` // pending, running, done, failed, cancelled
	Progress   float64 `json:"progress"`
	Error      string  `json:"error,omitempty"`
}

type Progress struct {
	File    string  `json:"file"`
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
	ETA     string  `json:"eta"`
	Status  string  `json:"status"`
}
