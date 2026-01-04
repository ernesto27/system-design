package render

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Image cache to avoid re-fetching on reflow
var imageCache = make(map[string]image.Image)

// renderTextFieldObjects creates canvas objects for input/textarea fields
func renderTextFieldObjects(x, y, width, height float64, value, placeholder string, isFocused bool) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	// Border - blue when focused, gray otherwise
	var borderColor color.Color
	if isFocused {
		borderColor = color.RGBA{0, 120, 215, 255}
	} else {
		borderColor = color.RGBA{180, 180, 180, 255}
	}
	border := canvas.NewRectangle(borderColor)
	border.Resize(fyne.NewSize(float32(width), float32(height)))
	border.Move(fyne.NewPos(float32(x), float32(y)))
	objects = append(objects, border)

	// White background (inset by 1px)
	bg := canvas.NewRectangle(color.White)
	bg.Resize(fyne.NewSize(float32(width-2), float32(height-2)))
	bg.Move(fyne.NewPos(float32(x+1), float32(y+1)))
	objects = append(objects, bg)

	// Show typed value or placeholder
	if value != "" {
		lines := strings.Split(value, "\n")
		lineHeight := float32(18)
		var lastLineWidth float32

		for i, line := range lines {
			text := canvas.NewText(line, color.Black)
			text.TextSize = 14
			text.Move(fyne.NewPos(float32(x+6), float32(y+6)+float32(i)*lineHeight))
			objects = append(objects, text)
			lastLineWidth = fyne.MeasureText(line, 14, fyne.TextStyle{}).Width
		}

		if isFocused {
			// Cursor at end of last line
			cursorY := float32(y+5) + float32(len(lines)-1)*lineHeight
			cursor := canvas.NewRectangle(color.Black)
			cursor.Resize(fyne.NewSize(1, 16))
			cursor.Move(fyne.NewPos(float32(x+6)+lastLineWidth, cursorY))
			objects = append(objects, cursor)
		}
	} else if placeholder != "" {
		text := canvas.NewText(placeholder, color.RGBA{150, 150, 150, 255})
		text.TextSize = 14
		text.Move(fyne.NewPos(float32(x+6), float32(y+6)))
		objects = append(objects, text)

		if isFocused {
			cursor := canvas.NewRectangle(color.Black)
			cursor.Resize(fyne.NewSize(1, 16))
			cursor.Move(fyne.NewPos(float32(x+6), float32(y+5)))
			objects = append(objects, cursor)
		}
	} else if isFocused {
		cursor := canvas.NewRectangle(color.Black)
		cursor.Resize(fyne.NewSize(1, 16))
		cursor.Move(fyne.NewPos(float32(x+6), float32(y+5)))
		objects = append(objects, cursor)
	}

	return objects
}

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
			text.TextStyle = fyne.TextStyle{
				Bold:      c.Bold,
				Italic:    c.Italic,
				Monospace: c.Monospace,
			}
			text.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, text)

			// Draw text decoration lines
			if c.Underline || c.Strikethrough {
				lineHeight := float32(1)
				var lineY float32
				if c.Underline {
					lineY = float32(c.Y) + c.Size + 2 // Below the text with small gap
				} else {
					lineY = float32(c.Y) + c.Size*0.4 // Through the middle
				}
				line := canvas.NewRectangle(c.Color)
				line.Resize(fyne.NewSize(float32(c.Width), lineHeight))
				line.Move(fyne.NewPos(float32(c.X), lineY))
				objects = append(objects, line)
			}

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

		case DrawInput:
			objects = append(objects, renderTextFieldObjects(c.X, c.Y, c.Width, c.Height, c.Value, c.Placeholder, c.IsFocused)...)

		case DrawButton:
			// Button background
			bg := canvas.NewRectangle(color.RGBA{225, 225, 225, 255})
			bg.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			bg.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, bg)

			// Top/left highlight
			highlight := canvas.NewRectangle(color.RGBA{255, 255, 255, 255})
			highlight.Resize(fyne.NewSize(float32(c.Width-1), 1))
			highlight.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, highlight)

			// Bottom/right shadow
			shadow := canvas.NewRectangle(color.RGBA{150, 150, 150, 255})
			shadow.Resize(fyne.NewSize(float32(c.Width), 1))
			shadow.Move(fyne.NewPos(float32(c.X), float32(c.Y+c.Height-1)))
			objects = append(objects, shadow)

			// Button text (centered)
			text := canvas.NewText(c.Text, color.Black)
			text.TextSize = 14
			textWidth := fyne.MeasureText(c.Text, 14, fyne.TextStyle{}).Width
			textX := c.X + (c.Width-float64(textWidth))/2
			text.Move(fyne.NewPos(float32(textX), float32(c.Y+8)))
			objects = append(objects, text)

		case DrawTextarea:
			objects = append(objects, renderTextFieldObjects(c.X, c.Y, c.Width, c.Height, c.Value, c.Placeholder, c.IsFocused)...)

		case DrawSelect:
			// Border
			border := canvas.NewRectangle(color.RGBA{180, 180, 180, 255})
			border.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			border.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, border)

			// White background
			bg := canvas.NewRectangle(color.White)
			bg.Resize(fyne.NewSize(float32(c.Width-2), float32(c.Height-2)))
			bg.Move(fyne.NewPos(float32(c.X+1), float32(c.Y+1)))
			objects = append(objects, bg)

			// Placeholder text
			text := canvas.NewText(c.Placeholder, color.RGBA{100, 100, 100, 255})
			text.TextSize = 14
			text.Move(fyne.NewPos(float32(c.X+6), float32(c.Y+6)))
			objects = append(objects, text)

			// Dropdown arrow ▼
			arrow := canvas.NewText("▼", color.RGBA{100, 100, 100, 255})
			arrow.TextSize = 10
			arrow.Move(fyne.NewPos(float32(c.X+c.Width-16), float32(c.Y+8)))
			objects = append(objects, arrow)
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
