package dom

import (
	"fmt"
	"strings"
)

func (n *Node) Print(indent int) {
	prefix := strings.Repeat("  ", indent)

	switch n.Type {
	case Document:
		fmt.Println(prefix + "#document")
	case Element:
		fmt.Printf("%s<%s>\n", prefix, n.TagName)
	case Text:
		fmt.Printf("%s\"%s\"\n", prefix, n.Text)
	}

	for _, child := range n.Children {
		child.Print(indent + 1)
	}

}
