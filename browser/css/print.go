package css

import "fmt"

func (s Stylesheet) Print() {
	for _, rule := range s.Rules {
		for i, sel := range rule.Selectors {
			if i > 0 {
				fmt.Print(", ")
			}

			if sel.TagName != "" {
				fmt.Print(sel.TagName)
			}

			if sel.ID != "" {
				fmt.Print("#" + sel.ID)
			}

			for _, class := range sel.Classes {
				fmt.Print("." + class)
			}
		}
		fmt.Print(" {")

		for _, decl := range rule.Declarations {
			fmt.Printf(" %s: %s;", decl.Property, decl.Value)
		}
		fmt.Println(" }")
	}
}
