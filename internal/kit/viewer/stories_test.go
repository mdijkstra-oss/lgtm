//go:build gallery

package viewer

import (
	"testing"

	"strigil/internal/kit/story/storytest"
)

func TestStories(t *testing.T) { storytest.Check(t, Stories()) }
