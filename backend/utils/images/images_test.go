package images

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"strings"
	"testing"

	"oneimg/backend/models"
)

func TestProcessMainImageKeepsPNGEncodingWhenWebPDisabled(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var source bytes.Buffer
	if err := png.Encode(&source, img); err != nil {
		t.Fatal(err)
	}

	service := &ImageService{}
	data, format, mimeType, err := service.processMainImage(
		source.Bytes(), img, "png", "image/png", int64(source.Len()),
		models.Settings{SaveWebp: false, OriginalImage: false},
	)
	if err != nil {
		t.Fatalf("processMainImage() error = %v", err)
	}
	if format != "png" || mimeType != "image/png" {
		t.Fatalf("got format=%q mime=%q", format, mimeType)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("result is not valid PNG: %v", err)
	}
}

func TestProcessMainImageConvertsHEICWhenOriginalImageIsRequested(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	service := &ImageService{}

	data, format, mimeType, err := service.processMainImage(
		[]byte("original heic bytes"), img, "heic", "image/heic", 19,
		models.Settings{SaveWebp: false, OriginalImage: true},
	)
	if err != nil {
		t.Fatalf("processMainImage() error = %v", err)
	}
	if format != "jpeg" || mimeType != "image/jpeg" {
		t.Fatalf("got format=%q mime=%q, want JPEG", format, mimeType)
	}
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("converted result is not valid JPEG: %v", err)
	}
}

func TestProcessImageUsesWebPExtensionForThumbnail(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	header := multipartFileHeader(t, "source.png", "image/png", data.Bytes())
	file, err := header.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	service := &ImageService{}
	processed, err := service.ProcessImage(file, header, models.Settings{
		SaveWebp:      false,
		Thumbnail:     true,
		WebpQuality:   85,
		OriginalImage: false,
	})
	if err != nil {
		t.Fatalf("ProcessImage() error = %v", err)
	}
	if !strings.HasSuffix(processed.ThumbnailName, ".webp") {
		t.Fatalf("thumbnail name %q does not use .webp", processed.ThumbnailName)
	}
	if got := detectImageMIME(processed.ThumbnailBytes); got != "image/webp" {
		t.Fatalf("thumbnail MIME = %q, want image/webp", got)
	}
}

func TestValidateImageAllowsMismatchedDeclaredTypeForValidImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var data bytes.Buffer
	if err := jpeg.Encode(&data, img, nil); err != nil {
		t.Fatal(err)
	}
	header := multipartFileHeader(t, "1000036933.png", "image/png", data.Bytes())

	service := &ImageService{}
	if err := service.ValidateImage(header, []string{"image/jpeg", "image/png"}, 1024*1024); err != nil {
		t.Fatalf("expected valid image to pass even if declared content-type differs: %v", err)
	}
}

func TestValidateImageRejectsUnsupportedDetectedType(t *testing.T) {
	header := multipartFileHeader(t, "fake.png", "image/png", []byte("not an image file"))

	service := &ImageService{}
	if err := service.ValidateImage(header, []string{"image/jpeg", "image/png"}, 1024*1024); err == nil {
		t.Fatal("non-image content unexpectedly validated")
	}
}

func TestValidateImageAcceptsHEICForConversion(t *testing.T) {
	data := fileTypeBox("mif1", "heic")
	header := multipartFileHeader(t, "photo.heic", "application/octet-stream", data)

	service := &ImageService{}
	if err := service.ValidateImage(header, []string{"image/jpeg", "image/png"}, 1024*1024); err != nil {
		t.Fatalf("expected detected HEIC to be accepted for conversion: %v", err)
	}
}

func TestValidateImageRejectsHEICSequence(t *testing.T) {
	data := fileTypeBox("hevc", "msf1")
	header := multipartFileHeader(t, "animation.heic", "image/heic-sequence", data)

	service := &ImageService{}
	if err := service.ValidateImage(header, []string{"image/jpeg", "image/png"}, 1024*1024); err == nil {
		t.Fatal("HEIC sequence unexpectedly validated")
	}
	if _, _, err := service.decodeImage(bytes.NewReader(data)); err == nil {
		t.Fatal("HEIC sequence unexpectedly decoded")
	}
}

func TestDetectImageMIMERecognizesHEIFBrands(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{name: "HEIC major brand", data: fileTypeBox("heic"), want: "image/heic"},
		{name: "HEIC compatible brand", data: fileTypeBox("mif1", "heic"), want: "image/heic"},
		{name: "generic HEIF", data: fileTypeBox("mif1"), want: "image/heif"},
		{name: "HEIC sequence", data: fileTypeBox("hevc"), want: "image/heic-sequence"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectImageMIME(tt.data); got != tt.want {
				t.Fatalf("detectImageMIME() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectImageMIMEDoesNotTreatAVIFAsHEIF(t *testing.T) {
	if got := detectImageMIME(fileTypeBox("avif", "mif1")); IsConvertibleImageMIME(got) {
		t.Fatalf("AVIF was misclassified as convertible HEIF: %q", got)
	}
}

func multipartFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader := multipart.NewReader(&body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		t.Fatal(err)
	}
	header := form.File["file"][0]
	header.Header.Set("Content-Type", contentType)
	return header
}

func fileTypeBox(majorBrand string, compatibleBrands ...string) []byte {
	size := 16 + len(compatibleBrands)*4
	data := make([]byte, size)
	binary.BigEndian.PutUint32(data[:4], uint32(size))
	copy(data[4:8], "ftyp")
	copy(data[8:12], majorBrand)
	for i, brand := range compatibleBrands {
		copy(data[16+i*4:20+i*4], brand)
	}
	return data
}
