package argonaut

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type ls struct {
	Command       CommandName `argonaut:"ls,joiner=[*]"`
	All           bool        `argonaut:"all|a"`
	LongFormat    bool        `argonaut:"l"`
	BlockSize     int         `argonaut:"block-size,long,joiner=[=]"`
	CoolStuff     string      `argonaut:"cool-stuff,long"`
	HumanReadable bool        `argonaut:"human-readable|h"`
	Paths         []string    `argonaut:",positional"`
}

func TestBasicMarshal(t *testing.T) {
	assert := require.New(t)

	output, err := Marshal(&ls{
		All:           true,
		LongFormat:    true,
		HumanReadable: true,
		BlockSize:     1024,
		CoolStuff:     `yep`,
		Paths: []string{
			`/foo`,
			`/bar/*.txt`,
			`/baz/`,
		},
	})

	assert.NoError(err)
	assert.Equal(`ls --all -l --block-size=1024 --cool-stuff*yep --human-readable /foo /bar/*.txt /baz/`, string(output))
	t.Logf(string(output))
}

// hateful complexity test 1: ffmpeg
// -------------------------------------------------------------------------------------------------
func TestFfmpegMarshal(t *testing.T) {
	assert := require.New(t)

	cmd := &FFMPEG{
		GlobalOptions: &GlobalOptions{
			LogLevel: `error`,
		},
		InputOptions: &InputOptions{
			URL: `/my/file.avi`,
		},
		OutputOptions: &OutputOptions{
			Common: Common{
				Codecs: []CodecOptions{
					{
						Stream: `v`,
						Codec:  `libx264`,
						Parameters: []string{
							`-preset`, `veryfast`,
							`-x264opts`, `keyint=24:min-keyint=24:scenecut=-1`,
							`-pix_fmt`, `yuv420p`,
						},
					}, {
						Stream: `a`,
						Codec:  `aac`,
					},
				},
			},
			URL: `/my/file.mkv`,
		},
	}

	output, err := Marshal(cmd)
	assert.NoError(err)

	should := `ffmpeg -loglevel error -i /my/file.avi -codec:v libx264 -preset veryfast -x264opts keyint=24:min-keyint=24:scenecut=-1 -pix_fmt yuv420p -codec:a aac /my/file.mkv`

	assert.Equal(should, string(output))
}
