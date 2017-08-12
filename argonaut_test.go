package argonaut

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Ls struct {
	All       bool     `argonaut:"all|a"`
	AlmostAll bool     `argonaut:"A"`
	BlockSize int      `argonaut:"block-size,long,required"`
	Format    string   `argonaut:"format,long,delimiters=[=]"`
	Paths     []string `argonaut:",positional"`
}

func TestBasicMarshal(t *testing.T) {
	assert := require.New(t)

	output, err := Marshal(&Ls{
		All:       true,
		AlmostAll: true,
		Format:    `across`,
		Paths: []string{
			`/foo`,
			`/bar/*.txt`,
			`/baz/`,
		},
	})

	assert.NoError(err)
	assert.Equal(`ls --all -A --block-size 0 --format=across /foo /bar/*.txt /baz/`, string(output))
}

// hateful complexity test 1: ffmpeg
// -------------------------------------------------------------------------------------------------
func TestFfmpegMarshal(t *testing.T) {
	assert := require.New(t)

	cmd := &FFMPEG{
		Global: &GlobalOptions{
			LogLevel: `error`,
		},
		Input: &InputOptions{
			URL: `/my/file.avi`,
		},
		Output: &OutputOptions{
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
