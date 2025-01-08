package cmdutil

import "github.com/spf13/cobra"

// AddGroup adds a group of commands to the parent command.
func AddGroup(parent *cobra.Command, title string, children ...*cobra.Command) {
	group := &cobra.Group{ID: title, Title: title}

	// Add add the group to the parent command.
	parent.AddGroup(group)

	// Add the child commands to the group.
	for _, child := range children {
		child.GroupID = title
		parent.AddCommand(child)
	}
}
