package models

// ImageMetadata is used to send some image parameters into
// metadata sql logic;
type ImageMetadata struct {
	OriginalFilename  string
	OriginalPath      string
	MimeType          string
	FileSize          int
	Status            string
	ProcessedVersions string
}
