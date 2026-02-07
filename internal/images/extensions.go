package images

import "strings"

type ImageExt string

const (
	ExtJPG  ImageExt = ".jpg"
	ExtJPEG ImageExt = ".jpeg"
	ExtPNG  ImageExt = ".png"
	ExtGIF  ImageExt = ".gif"
	ExtWEBP ImageExt = ".webp"
	ExtBMP  ImageExt = ".bmp"
	ExtTIF  ImageExt = ".tif"
	ExtTIFF ImageExt = ".tiff"
)

// ExtSet fast lookup set
var extSet = map[string]ImageExt{
	string(ExtJPG):  ExtJPG,
	string(ExtJPEG): ExtJPEG,
	string(ExtPNG):  ExtPNG,
	string(ExtGIF):  ExtGIF,
	string(ExtWEBP): ExtWEBP,
	string(ExtBMP):  ExtBMP,
	string(ExtTIF):  ExtTIF,
	string(ExtTIFF): ExtTIFF,
}

// IsSupported returns true if the extension is supported
func IsSupported(ext string) bool {
	ext = strings.ToLower(ext)
	_, ok := extSet[ext]
	return ok
}
