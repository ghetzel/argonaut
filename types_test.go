package argonaut

type TimeDuration string
type DateSpecification string

type GlobalOptions struct {
	ForceOverwrite       bool     `argonaut:"y"`
	NeverOverwrite       bool     `argonaut:"n"`
	ShowHelp             bool     `argonaut:"help|h|?"`
	ShowHelpSection      string   `argonaut:"help,long"`
	ShowLicense          bool     `argonaut:"L"`
	ShowVersion          bool     `argonaut:"version"`
	ListFormats          bool     `argonaut:"formats"`
	ListDevices          bool     `argonaut:"devices"`
	ListCodecs           bool     `argonaut:"codecs"`
	ListDecoders         bool     `argonaut:"decoders"`
	ListEncoders         bool     `argonaut:"encoders"`
	ListBitstreamFilters bool     `argonaut:"bsfs"`
	ListProtocols        bool     `argonaut:"protocols"`
	ListFilters          bool     `argonaut:"filters"`
	ListPixelFormats     bool     `argonaut:"pix_fmts"`
	ListSampleFormats    bool     `argonaut:"sample_fmts"`
	ListLayouts          bool     `argonaut:"layouts"`
	ListColors           bool     `argonaut:"colors"`
	LogLevel             string   `argonaut:"loglevel|v,short"`
	DumpReport           bool     `argonaut:"report"`
	HideBanner           bool     `argonaut:"hide_banner"`
	CpuFlags             []string `argonaut:"cpuflags"`
}

type CodecOptions struct {
	ArgName    ArgonautArgument `argonaut:"codec,short"`
	Stream     string           `argonaut:",suffixprev,delimiters=[:]"`
	Codec      string           `argonaut:",skipname"`
	Parameters []string         `argonaut:",positional"`
}

type MetadataValue struct {
	Metadata   ArgonautArgument `argonaut:",short"`
	Metastream string           `argonaut:",suffixprev,delimiters=[:]"`
	Key        string
	Value      interface{}
}

type Common struct {
	Format    string `argonaut:"f"`
	Codecs    []CodecOptions
	Duration  TimeDuration `argonaut:"t"`
	SeekStart TimeDuration `argonaut:"ss"`
	SeekEnd   TimeDuration `argonaut:"sseof"`
}

type InputOptions struct {
	Common
	InputTimeOffset TimeDuration `argonaut:"itsoffset"`
	Metadata        []MetadataValue
	URL             string `argonaut:"i,required"`
}

type OutputOptions struct {
	Common
	OutputDuration TimeDuration      `argonaut:"to"`
	Timestamp      DateSpecification `argonaut:"timestamp"`
	LimitSize      int64             `argonaut:"fs"`
	URL            string            `argonaut:",positional,required"`
}

type FFMPEG struct {
	Command ArgonautCommand `argonaut:"ffmpeg"`
	Global  *GlobalOptions  `argonaut:",label=global_options"`
	Input   *InputOptions   `argonaut:",label=input_file_options"`
	Output  *OutputOptions  `argonaut:",label=output_file_options"`
}
