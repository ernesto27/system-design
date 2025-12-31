package render

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Image cache to avoid re-fetching on reflow
var imageCache = make(map[string]image.Image)

func RenderToCanvas(commands []DisplayCommand, baseURL string, useCache bool) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	for _, cmd := range commands {
		switch c := cmd.(type) {
		case DrawRect:
			rect := canvas.NewRectangle(c.Color)
			rect.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			rect.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, rect)

		case DrawText:
			text := canvas.NewText(c.Text, c.Color)
			text.TextSize = c.Size
			text.TextStyle = fyne.TextStyle{Bold: c.Bold, Italic: c.Italic}
			text.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, text)

		case DrawImage:
			var img *canvas.Image
			if useCache {
				img = createImageFromCache(c.URL, baseURL, c.Width, c.Height)
			} else {
				img = fetchAndCreateImage(c.URL, baseURL, c.Width, c.Height)
			}
			if img != nil {
				img.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
				objects = append(objects, img)
			}

		case DrawHR:
			hr := canvas.NewRectangle(color.Gray{Y: 180})
			hr.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			hr.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, hr)
		}
	}

	return objects
}

func fetchAndCreateImage(src, baseURL string, width, height float64) *canvas.Image {
	fullURL := resolveImageURL(src, baseURL)
	fmt.Println("Fetching image:", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error fetching image:", err)
		return nil
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return nil
	}

	// Cache the image
	imageCache[fullURL] = img

	fyneImg := canvas.NewImageFromImage(img)
	fyneImg.FillMode = canvas.ImageFillContain
	fyneImg.Resize(fyne.NewSize(float32(width), float32(height)))

	return fyneImg
}

// createImageFromCache uses cached image data
func createImageFromCache(src, baseURL string, width, height float64) *canvas.Image {
	fullURL := resolveImageURL(src, baseURL)

	// Check cache first
	if cached, ok := imageCache[fullURL]; ok {
		fyneImg := canvas.NewImageFromImage(cached)
		fyneImg.FillMode = canvas.ImageFillContain
		fyneImg.Resize(fyne.NewSize(float32(width), float32(height)))
		return fyneImg
	}

	// Not cached, fetch it
	return fetchAndCreateImage(src, baseURL, width, height)
}

func resolveImageURL(src, baseURL string) string {
	if len(src) > 4 && (src[:4] == "http" || src[:2] == "//") {
		if src[:2] == "//" {
			return "https:" + src
		}
		return src
	}

	if baseURL == "" {
		return src
	}

	if len(src) > 0 && src[0] == '/' {
		return baseURL + src
	}

	return baseURL + "/" + src
}
