package ui

import "sync/atomic"

var (
	Height atomic.Int32
	Width  atomic.Int32
)

func SetHeight(h int) {
	Height.Store(int32(h))
}
func GetHeight() int {
	return int(Height.Load())
}

func SetWidth(w int) {
	Width.Store(int32(w))
}
func GetWidth() int {
	return int(Width.Load())
}
