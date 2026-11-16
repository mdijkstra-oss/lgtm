//go:build gallery

package palette

import (
	"testing"

	"strigil/internal/kit/story/storytest"
	"strigil/internal/kit/theme"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories(theme.Mocha())) }
