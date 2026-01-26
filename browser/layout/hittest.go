package layout

func (box *LayoutBox) HitTest(x, y float64) *LayoutBox {
	if !box.Contains(x, y) {
		return nil
	}

	for i := len(box.Children) - 1; i >= 0; i-- {
		child := box.Children[i]
		if hit := child.HitTest(x, y); hit != nil {
			return hit
		}
	}

	return box
}

// Contains checks if point (x, y) is inside this box
func (box *LayoutBox) Contains(x, y float64) bool {
	return x >= box.Rect.X &&
		x <= box.Rect.X+box.Rect.Width &&
		y >= box.Rect.Y &&
		y <= box.Rect.Y+box.Rect.Height
}

// FindLink walks up the tree to find an <a> element
func (box *LayoutBox) FindLink() string {
	current := box

	for current != nil {
		if current.Node != nil && current.Node.TagName == "a" {
			if href, ok := current.Node.Attributes["href"]; ok {
				return href
			}
		}
		// Walk up to parent (we need to add parent tracking)
		current = current.Parent
	}

	return ""
}

type LinkInfo struct {
	Href   string
	Target string
	Rel    string
}

func (box *LayoutBox) FindLinkInfo() *LinkInfo {
	current := box

	for current != nil {
		if current.Node != nil && current.Node.TagName == "a" {
			href, hasHref := current.Node.Attributes["href"]
			if hasHref {
				return &LinkInfo{
					Href:   href,
					Target: current.Node.Attributes["target"],
					Rel:    current.Node.Attributes["rel"],
				}
			}
		}
		current = current.Parent
	}

	return nil
}
