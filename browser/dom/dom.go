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
