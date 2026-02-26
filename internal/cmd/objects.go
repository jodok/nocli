package cmd

import (
	"fmt"
)

type ObjectsCmd struct{}

func (c *ObjectsCmd) Run() error {
	fmt.Println("Available object entry points:")
	fmt.Println("  nocli page objects <page-url-or-id>      # Flatten page recordMap tables")
	fmt.Println("  nocli page objects <page> --table block --notion-block-like")
	fmt.Println("                                            # Notion-like block objects")
	fmt.Println("  nocli page types <page-url-or-id>         # Seen block types vs public API types")
	fmt.Println("  nocli block get <block-id> --notion-block-like")
	fmt.Println("                                            # Single block as normalized object")
	fmt.Println("  nocli block children <block-id>           # Direct child block objects")
	fmt.Println("  nocli collection query <collection-id> <view-id> --flatten")
	fmt.Println("                                            # Collection/view object rows")
	return nil
}
