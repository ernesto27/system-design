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
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	imageCache      = make(map[string]image.Image)
	imageCacheMu    sync.Mutex
	pendingFeteches = make(map[string]bool)
	pendingMu       sync.Mutex
)

// renderTextFieldObjects creates canvas objects for input/textarea fields
func renderTextFieldObjects(x, y, width, height float64, value, placeholder string, isFocused, isDisabled, isValid bool) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	// Border color based on state
	var borderColor color.Color
	if isDisabled {
		borderColor = ColorBorderDisabled
	} else if !isValid {
		borderColor = ColorBorderInvalid
	} else if isFocused {
		borderColor = ColorBorderFocused
	} else {
		borderColor = ColorBorder
	}
	border := canvas.NewRectangle(borderColor)
	border.Resize(fyne.NewSize(float32(width), float32(height)))
	border.Move(fyne.NewPos(float32(x), float32(y)))
	objects = append(objects, border)

	// Background (inset by 1px)
	bgColor := ColorInputBg
	if isDisabled {
		bgColor = ColorInputBgDisabled
	}
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(float32(width-2), float32(height-2)))
	bg.Move(fyne.NewPos(float32(x+1), float32(y+1)))
	objects = append(objects, bg)

	// Text color based on state
	textColor := ColorText
	if isDisabled {
		textColor = ColorTextDisabled
	}

	// Show typed value or placeholder
	if value != "" {
		lines := strings.Split(value, "\n")
		lineHeight := float32(18)
		var lastLineWidth float32

		for i, line := range lines {
			text := canvas.NewText(line, textColor)
			text.TextSize = 14
			text.Move(fyne.NewPos(float32(x+6), float32(y+6)+float32(i)*lineHeight))
			objects = append(objects, text)
			lastLineWidth = fyne.MeasureText(line, 14, fyne.TextStyle{}).Width
		}

		if isFocused && !isDisabled {
			cursorY := float32(y+5) + float32(len(lines)-1)*lineHeight
			cursor := canvas.NewRectangle(ColorBlack)
			cursor.Resize(fyne.NewSize(1, 16))
			cursor.Move(fyne.NewPos(float32(x+6)+lastLineWidth, cursorY))
			objects = append(objects, cursor)
		}
	} else if placeholder != "" {
		placeholderColor := ColorPlaceholder
		if isDisabled {
			placeholderColor = ColorPlaceholderDisabled
		}

		text := canvas.NewText(placeholder, placeholderColor)
		text.TextSize = 14
		text.Move(fyne.NewPos(float32(x+6), float32(y+6)))
		objects = append(objects, text)

		if isFocused && !isDisabled {
			cursor := canvas.NewRectangle(ColorBlack)
			cursor.Resize(fyne.NewSize(1, 16))
			cursor.Move(fyne.NewPos(float32(x+6), float32(y+5)))
			objects = append(objects, cursor)
		}
	} else if isFocused && !isDisabled {
		cursor := canvas.NewRectangle(ColorBlack)
		cursor.Resize(fyne.NewSize(1, 16))
		cursor.Move(fyne.NewPos(float32(x+6), float32(y+5)))
		objects = append(objects, cursor)
	}

	return objects
}

// renderNumberInput creates canvas objects for number input with spin buttons
func renderNumberInput(x, y, width, height float64, value, placeholder string, isFocused, isDisabled bool) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	buttonWidth := 24.0
	textFieldWidth := width - buttonWidth

	// Border color based on state
	var borderColor color.Color
	if isDisabled {
		borderColor = ColorBorderDisabled
	} else if isFocused {
		borderColor = ColorBorderFocused
	} else {
		borderColor = ColorBorder
	}

	// Main border around entire control
	border := canvas.NewRectangle(borderColor)
	border.Resize(fyne.NewSize(float32(width), float32(height)))
	border.Move(fyne.NewPos(float32(x), float32(y)))
	objects = append(objects, border)

	// Text field background
	bgColor := ColorInputBg
	if isDisabled {
		bgColor = ColorInputBgDisabled
	}
	bg := canvas.NewRectangle(bgColor)
	bg.Resize(fyne.NewSize(float32(textFieldWidth-2), float32(height-2)))
	bg.Move(fyne.NewPos(float32(x+1), float32(y+1)))
	objects = append(objects, bg)

	// Text color
	textColor := ColorText
	if isDisabled {
		textColor = ColorTextDisabled
	}

	// Display value or placeholder
	if value != "" {
		text := canvas.NewText(value, textColor)
		text.TextSize = 14
		text.Move(fyne.NewPos(float32(x+6), float32(y+6)))
		objects = append(objects, text)

		if isFocused && !isDisabled {
			textWidth := fyne.MeasureText(value, 14, fyne.TextStyle{}).Width
			cursor := canvas.NewRectangle(ColorBlack)
			cursor.Resize(fyne.NewSize(1, 16))
			cursor.Move(fyne.NewPos(float32(x+6)+textWidth, float32(y+5)))
			objects = append(objects, cursor)
		}
	} else if placeholder != "" {
		placeholderColor := ColorPlaceholder
		if isDisabled {
			placeholderColor = ColorPlaceholderDisabled
		}
		text := canvas.NewText(placeholder, placeholderColor)
		text.TextSize = 14
		text.Move(fyne.NewPos(float32(x+6), float32(y+6)))
		objects = append(objects, text)
	}

	// Spin buttons area
	btnX := x + textFieldWidth
	btnHeight := height / 2

	// Button colors
	btnBg := ColorButtonBg
	btnText := ColorText
	if isDisabled {
		btnBg = ColorButtonBgDisabled
		btnText = ColorTextDisabled
	}

	// Up button (top half)
	upBg := canvas.NewRectangle(btnBg)
	upBg.Resize(fyne.NewSize(float32(buttonWidth-1), float32(btnHeight-1)))
	upBg.Move(fyne.NewPos(float32(btnX), float32(y+1)))
	objects = append(objects, upBg)

	upArrow := canvas.NewText("▲", btnText)
	upArrow.TextSize = 10
	upArrow.Move(fyne.NewPos(float32(btnX+7), float32(y+2)))
	objects = append(objects, upArrow)

	// Down button (bottom half)
	downBg := canvas.NewRectangle(btnBg)
	downBg.Resize(fyne.NewSize(float32(buttonWidth-1), float32(btnHeight-1)))
	downBg.Move(fyne.NewPos(float32(btnX), float32(y+btnHeight)))
	objects = append(objects, downBg)

	downArrow := canvas.NewText("▼", btnText)
	downArrow.TextSize = 10
	downArrow.Move(fyne.NewPos(float32(btnX+7), float32(y+btnHeight+1)))
	objects = append(objects, downArrow)

	// Separator line between buttons
	separator := canvas.NewRectangle(borderColor)
	separator.Resize(fyne.NewSize(float32(buttonWidth-1), 1))
	separator.Move(fyne.NewPos(float32(btnX), float32(y+btnHeight)))
	objects = append(objects, separator)

	return objects
}

func RenderToCanvas(commands []DisplayCommand, baseURL string, useCache bool, onImageLoad func()) []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	var dropdownOverlays []fyne.CanvasObject // Collect dropdowns to render LAST (on top)

	for _, cmd := range commands {
		switch c := cmd.(type) {
		case DrawRect:
			rect := canvas.NewRectangle(c.Color)
			rect.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			rect.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			rect.CornerRadius = float32(c.CornerRadius)
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
					lineY = float32(c.Y) + c.Size + 2
				} else {
					lineY = float32(c.Y) + c.Size*0.9
				}
				line := canvas.NewRectangle(c.Color)
				line.Resize(fyne.NewSize(float32(c.Width), lineHeight))
				line.Move(fyne.NewPos(float32(c.X), lineY))
				objects = append(objects, line)
			}

		case DrawImage:
			img := getImageOrPlaceholder(c.URL, baseURL, c.Width, c.Height, onImageLoad)

			if img != nil {
				img.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
				objects = append(objects, img)
			} else {
				// Not cached yet - show gray placeholder
				placeholder := canvas.NewRectangle(color.RGBA{220, 220, 220, 255})
				placeholder.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
				placeholder.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
				objects = append(objects, placeholder)
			}

		case DrawHR:
			hr := canvas.NewRectangle(ColorHR)
			hr.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			hr.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, hr)

		case DrawInput:
			displayValue := c.Value
			if c.InputType == "password" && displayValue != "" {
				displayValue = strings.Repeat("•", len([]rune(displayValue)))
			}
			if c.InputType == "number" {
				objects = append(objects, renderNumberInput(c.X, c.Y, c.Width, c.Height, displayValue, c.Placeholder, c.IsFocused, c.IsDisabled)...)
			} else {
				objects = append(objects, renderTextFieldObjects(c.X, c.Y, c.Width, c.Height, displayValue, c.Placeholder, c.IsFocused, c.IsDisabled, c.IsValid)...)
			}

		case DrawButton:
			// Button background
			bgColor := ColorButtonBg
			if c.IsDisabled {
				bgColor = ColorButtonBgDisabled
			}
			bg := canvas.NewRectangle(bgColor)
			bg.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			bg.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, bg)

			// Top/left highlight
			highlightColor := ColorButtonHighlight
			if c.IsDisabled {
				highlightColor = ColorButtonHighlightDisabled
			}
			highlight := canvas.NewRectangle(highlightColor)
			highlight.Resize(fyne.NewSize(float32(c.Width-1), 1))
			highlight.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, highlight)

			// Bottom/right shadow
			shadowColor := ColorButtonShadow
			if c.IsDisabled {
				shadowColor = ColorButtonShadowDisabled
			}
			shadow := canvas.NewRectangle(shadowColor)
			shadow.Resize(fyne.NewSize(float32(c.Width), 1))
			shadow.Move(fyne.NewPos(float32(c.X), float32(c.Y+c.Height-1)))
			objects = append(objects, shadow)

			// Button text (centered)
			textColor := ColorText
			if c.IsDisabled {
				textColor = ColorTextDisabled
			}
			text := canvas.NewText(c.Text, textColor)
			text.TextSize = 14
			textWidth := fyne.MeasureText(c.Text, 14, fyne.TextStyle{}).Width
			textX := c.X + (c.Width-float64(textWidth))/2
			text.Move(fyne.NewPos(float32(textX), float32(c.Y+8)))
			objects = append(objects, text)

		case DrawTextarea:
			objects = append(objects, renderTextFieldObjects(c.X, c.Y, c.Width, c.Height, c.Value, c.Placeholder, c.IsFocused, c.IsDisabled, true)...)

		case DrawSelect:
			// Border - blue when open
			borderColor := ColorBorder
			if c.IsOpen {
				borderColor = ColorBorderFocused
			}
			if c.IsDisabled {
				borderColor = ColorBorderDisabled
			}
			border := canvas.NewRectangle(borderColor)
			border.Resize(fyne.NewSize(float32(c.Width), float32(c.Height)))
			border.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, border)

			// Background
			bgColor := ColorInputBg
			if c.IsDisabled {
				bgColor = ColorInputBgDisabled
			}
			bg := canvas.NewRectangle(bgColor)
			bg.Resize(fyne.NewSize(float32(c.Width-2), float32(c.Height-2)))
			bg.Move(fyne.NewPos(float32(c.X+1), float32(c.Y+1)))
			objects = append(objects, bg)

			// Selected value or placeholder
			displayText := "Select..."
			textColor := ColorSelectArrow
			if c.SelectedValue != "" {
				displayText = c.SelectedValue
				textColor = ColorText
			}
			if c.IsDisabled {
				textColor = ColorTextDisabled
			}
			text := canvas.NewText(displayText, textColor)
			text.TextSize = 14
			text.Move(fyne.NewPos(float32(c.X+6), float32(c.Y+6)))
			objects = append(objects, text)

			// Dropdown arrow
			arrowText := "▼"
			if c.IsOpen {
				arrowText = "▲"
			}
			arrowColor := ColorSelectArrow
			if c.IsDisabled {
				arrowColor = ColorTextDisabled
			}
			arrow := canvas.NewText(arrowText, arrowColor)
			arrow.TextSize = 10
			arrow.Move(fyne.NewPos(float32(c.X+c.Width-16), float32(c.Y+8)))
			objects = append(objects, arrow)

			// Dropdown list when open
			if c.IsOpen && len(c.Options) > 0 {
				fmt.Printf("Canvas: Rendering dropdown with %d options at Y=%.0f\n", len(c.Options), c.Y+c.Height)
				optionHeight := float64(28)
				dropdownHeight := optionHeight * float64(len(c.Options))

				// Dropdown border
				dropBorder := canvas.NewRectangle(ColorBorder)
				dropBorder.Resize(fyne.NewSize(float32(c.Width), float32(dropdownHeight+2)))
				dropBorder.Move(fyne.NewPos(float32(c.X), float32(c.Y+c.Height)))
				dropdownOverlays = append(dropdownOverlays, dropBorder)

				// Dropdown background
				dropBg := canvas.NewRectangle(ColorWhite)
				dropBg.Resize(fyne.NewSize(float32(c.Width-2), float32(dropdownHeight)))
				dropBg.Move(fyne.NewPos(float32(c.X+1), float32(c.Y+c.Height+1)))
				dropdownOverlays = append(dropdownOverlays, dropBg)

				// Options
				for i, opt := range c.Options {
					optY := c.Y + c.Height + float64(i)*optionHeight

					// Highlight selected option
					if opt == c.SelectedValue {
						highlight := canvas.NewRectangle(ColorSelectHighlight)
						highlight.Resize(fyne.NewSize(float32(c.Width-2), float32(optionHeight)))
						highlight.Move(fyne.NewPos(float32(c.X+1), float32(optY+1)))
						dropdownOverlays = append(dropdownOverlays, highlight)
					}

					optText := canvas.NewText(opt, ColorBlack)
					optText.TextSize = 14
					optText.Move(fyne.NewPos(float32(c.X+6), float32(optY+6)))
					dropdownOverlays = append(dropdownOverlays, optText)
				}
			}

		case DrawRadio:
			size := float32(c.Width)
			if float32(c.Height) < size {
				size = float32(c.Height)
			}

			// Outer circle
			outerColor := ColorCheckboxBorder
			if c.IsDisabled {
				outerColor = ColorCheckboxBorderDisabled
			}
			outerCircle := canvas.NewCircle(outerColor)
			outerCircle.Resize(fyne.NewSize(size, size))
			outerCircle.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, outerCircle)

			// Inner background
			innerSize := size - 4
			innerColor := ColorInputBg
			if c.IsDisabled {
				innerColor = ColorInputBgDisabled
			}
			innerCircle := canvas.NewCircle(innerColor)
			innerCircle.Resize(fyne.NewSize(innerSize, innerSize))
			innerCircle.Move(fyne.NewPos(float32(c.X)+2, float32(c.Y)+2))
			objects = append(objects, innerCircle)

			if c.IsChecked {
				dotSize := size - 10
				dotColor := ColorAccent
				if c.IsDisabled {
					dotColor = ColorAccentDisabled
				}
				dot := canvas.NewCircle(dotColor)
				dot.Resize(fyne.NewSize(dotSize, dotSize))
				dot.Move(fyne.NewPos(float32(c.X)+5, float32(c.Y)+5))
				objects = append(objects, dot)
			}

		case DrawCheckbox:
			size := float32(c.Width)
			if float32(c.Height) < size {
				size = float32(c.Height)
			}

			// Border
			borderColor := ColorCheckboxBorder
			if c.IsDisabled {
				borderColor = ColorCheckboxBorderDisabled
			}
			border := canvas.NewRectangle(borderColor)
			border.Resize(fyne.NewSize(size, size))
			border.Move(fyne.NewPos(float32(c.X), float32(c.Y)))
			objects = append(objects, border)

			// Inner background
			innerSize := size - 4
			innerColor := ColorInputBg
			if c.IsDisabled {
				innerColor = ColorInputBgDisabled
			}
			inner := canvas.NewRectangle(innerColor)
			inner.Resize(fyne.NewSize(innerSize, innerSize))
			inner.Move(fyne.NewPos(float32(c.X)+2, float32(c.Y)+2))
			objects = append(objects, inner)

			if c.IsChecked {
				checkColor := ColorAccent
				if c.IsDisabled {
					checkColor = ColorAccentDisabled
				}
				check := canvas.NewText("✓", checkColor)
				check.TextSize = size - 6
				check.Move(fyne.NewPos(float32(c.X)+3, float32(c.Y)+1))
				objects = append(objects, check)
			}
		case DrawFileInput:
			objects = append(objects, renderFileInput(c.X, c.Y, c.Width, c.Height, c.Filename, c.IsDisabled)...)
		}
	}

	// Append dropdown overlays at the end so they render on top of everything
	objects = append(objects, dropdownOverlays...)

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

func renderFileInput(x, y, width, height float64, filename string, isDisabled bool) []fyne.CanvasObject {
	var objects []fyne.CanvasObject

	buttonWidth := 100.0

	// Button background
	btnBg := ColorButtonBg
	if isDisabled {
		btnBg = ColorButtonBgDisabled
	}
	btn := canvas.NewRectangle(btnBg)
	btn.Resize(fyne.NewSize(float32(buttonWidth), float32(height)))
	btn.Move(fyne.NewPos(float32(x), float32(y)))
	objects = append(objects, btn)

	// Button text
	btnTextColor := ColorText
	if isDisabled {
		btnTextColor = ColorTextDisabled
	}
	btnText := canvas.NewText("Choose File", btnTextColor)
	btnText.TextSize = 12
	btnText.Move(fyne.NewPos(float32(x+10), float32(y+8)))
	objects = append(objects, btnText)

	// Filename area background
	filenameBg := canvas.NewRectangle(ColorInputBg)
	filenameBg.Resize(fyne.NewSize(float32(width-buttonWidth-4), float32(height)))
	filenameBg.Move(fyne.NewPos(float32(x+buttonWidth+4), float32(y)))
	objects = append(objects, filenameBg)

	// Filename text
	displayName := "No file chosen"
	if filename != "" {
		// Show just the filename, not full path
		parts := strings.Split(filename, "/")
		displayName = parts[len(parts)-1]
	}
	filenameText := canvas.NewText(displayName, ColorText)
	filenameText.TextSize = 12
	filenameText.Move(fyne.NewPos(float32(x+buttonWidth+10), float32(y+8)))
	objects = append(objects, filenameText)

	return objects
}

func fetchimageToCache(fullURL string) {
	fmt.Println("Fetching image ", fullURL)
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("Error fetching image:", err)
		return
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	imageCacheMu.Lock()
	imageCache[fullURL] = img
	imageCacheMu.Unlock()
}

func getImageOrPlaceholder(src, baseURL string, width, height float64, onLoad func()) *canvas.Image {
	fullURL := resolveImageURL(src, baseURL)

	imageCacheMu.Lock()
	cached, found := imageCache[fullURL]
	imageCacheMu.Unlock()

	if found {
		fyneImg := canvas.NewImageFromImage(cached)
		fyneImg.Resize(fyne.NewSize(float32(width), float32(height)))
		fyneImg.FillMode = canvas.ImageFillContain
		return fyneImg
	}

	pendingMu.Lock()
	alreadyFetching := pendingFeteches[fullURL]
	if !alreadyFetching {
		pendingFeteches[fullURL] = true
	}
	pendingMu.Unlock()

	if !alreadyFetching {
		go func() {
			fetchimageToCache(fullURL)

			pendingMu.Lock()
			delete(pendingFeteches, fullURL)
			pendingMu.Unlock()

			if onLoad != nil {
				fyne.Do(onLoad)
			}
		}()
	}

	return nil
}
