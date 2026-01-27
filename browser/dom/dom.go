package dom

type NodeType int

const (
	Document NodeType = iota
	Element
	Text
)

type Node struct {
	Type       NodeType
	TagName    string
	Namespace  string
	Attributes map[string]string
	Children   []*Node
	Parent     *Node
	Text       string
}

func NewElement(tagName string, tags map[string]string) *Node {
	return &Node{
		Type:       Element,
		TagName:    tagName,
		Attributes: tags,
		Children:   []*Node{},
	}
}

func NewText(text string) *Node {
	return &Node{
		Type: Text,
		Text: text,
	}
}

func (n *Node) AppendChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *Node) RemoveChild(child *Node) {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return
		}
	}
}

func (n *Node) Remove() {
	if n.Parent != nil {
		n.Parent.RemoveChild(n)
	}
}

func FindTitle(node *Node) string {
	if node == nil {
		return ""
	}

	if node.TagName == TagTitle {
		for _, child := range node.Children {
			if child.Type == Text {
				return child.Text
			}
		}
	}

	for _, child := range node.Children {
		if title := FindTitle(child); title != "" {
			return title
		}
	}

	return ""
}

func FindStyleContent(node *Node) string {
	if node == nil {
		return ""
	}

	var css string

	if node.TagName == "style" {
		for _, child := range node.Children {
			if child.Type == Text {
				css += child.Text + "\n"
			}
		}
	}

	for _, child := range node.Children {
		css += FindStyleContent(child)
	}

	return css
}

func FindStylesheetLinks(node *Node) []string {
	var links []string
	if node.TagName == "link" && node.Attributes["rel"] == "stylesheet" {
		if href, ok := node.Attributes["href"]; ok {
			links = append(links, href)
		}
	}
	for _, child := range node.Children {
		links = append(links, FindStylesheetLinks(child)...)
	}
	return links
}

func FindElementsByTagName(node *Node, tagName string) *Node {
	if node == nil {
		return nil
	}

	if node.Type == Element && node.TagName == tagName {
		return node
	}

	for _, child := range node.Children {
		if found := FindElementsByTagName(child, tagName); found != nil {
			return found
		}
	}

	return nil
}

func FindBaseHref(node *Node) string {
	baseNode := FindElementsByTagName(node, TagBase)
	if baseNode == nil {
		return ""
	}
	return baseNode.Attributes["href"]
}
