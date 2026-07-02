package images

import (
	"bytes"
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

func TestValidateImageRejectsDeclaredTypeMismatch(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var data bytes.Buffer
	if err := jpeg.Encode(&data, img, nil); err != nil {
		t.Fatal(err)
	}
	header := multipartFileHeader(t, "spoofed.png", "image/png", data.Bytes())

	service := &ImageService{}
	if err := service.ValidateImage(header, []string{"image/jpeg", "image/png"}, 1024*1024); err == nil {
		t.Fatal("mismatched image content type unexpectedly validated")
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
