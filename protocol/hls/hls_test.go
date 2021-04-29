package hls

import (
	"path"
	"strings"
	"testing"
)

func TestPath(t *testing.T) {
	pathstr := strings.TrimLeft("/aaa/bbb.m3u8", "/")
	key := strings.Split(pathstr, path.Ext(pathstr))[0]
	t.Log(key)

	paths := strings.SplitN(pathstr, "/", 3)
	for i := range paths {
		t.Log(paths[i])
	}
}
