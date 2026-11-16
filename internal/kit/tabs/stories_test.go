//go:build gallery

package tabs

import (
	"testing"

	"strigil/internal/kit/story/storytest"
	"strigil/internal/kit/theme"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories(theme.Mocha())) }
