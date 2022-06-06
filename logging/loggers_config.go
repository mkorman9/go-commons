package logging

type LoggerConfig struct {
	Output string      `json:"output" yaml:"output" mapstructure:"output"`
	Gelf   *GelfConfig `json:"gelf" yaml:"gelf" mapstructure:"gelf"`
	File   *FileConfig `json:"file" yaml:"file" mapstructure:"file"`
	Format string      `json:"format" yaml:"format" mapstructure:"format"`
	Text   *TextConfig `json:"text" yaml:"text" mapstructure:"text"`
}

type GelfConfig struct {
	Address string `json:"address" yaml:"address" mapstructure:"address"`
}

type FileConfig struct {
	Location string `json:"location" yaml:"location" mapstructure:"location"`
}

type TextConfig struct {
	Colors bool `json:"colors" yaml:"colors" mapstructure:"colors"`
}
