package logging

type LoggerConfig struct {
	Output string      `json:"output" yaml:"output"`
	Gelf   *GelfConfig `json:"gelf" yaml:"gelf"`
	File   *FileConfig `json:"file" yaml:"file"`
	Format string      `json:"format" yaml:"format"`
	Text   *TextConfig `json:"text" yaml:"text"`
}

type GelfConfig struct {
	Address string `json:"address" yaml:"address"`
}

type FileConfig struct {
	Location string `json:"location" yaml:"location"`
}

type TextConfig struct {
	Colors bool `json:"colors" yaml:"colors"`
}
