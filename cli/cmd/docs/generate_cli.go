package docs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func GenerateCliDocsCmd(rootCmd *cobra.Command, ch *cmdutil.Helper) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:    "generate-cli",
		Short:  "Generate CLI documentation",
		Args:   cobra.ExactArgs(1),
		Hidden: !ch.IsDev(),
		Run: func(cmd *cobra.Command, args []string) {
			dir := args[0]
			rootCmd.DisableAutoGenTag = true
			err := genMarkdownTree(rootCmd, dir)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	return docsCmd
}

func genMarkdownTree(cmd *cobra.Command, dir string) error {
	emptyStr := func(s string) string { return "" }
	return genMarkdownTreeCustom(cmd, dir, emptyStr)
}

func genMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender func(string) string) error {
	if cmd.Hidden {
		return nil
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}

		sd := dir
		if cmd.Parent() != nil {
			sd = filepath.Join(dir, cmd.Name())
		}

		if err := genMarkdownTreeCustom(c, sd, filePrepender); err != nil {
			return err
		}
	}

	nm := cmd.Name()
	if cmd.Parent() == nil {
		nm = "cli"
	}

	filename := filepath.Join(dir, nm+".md")
	if len(cmd.Commands()) > 0 && cmd.Parent() != nil {
		dir = filepath.Join(dir, cmd.Name())
		filename = filepath.Join(dir, nm+".md")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, fs.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.WriteString(filePrepender(filename)); err != nil {
		return err
	}

	return genMarkdownCustom(cmd, f)
}

func genMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	/*
		---
		title: CLI cheat sheet
		sidebar_position: 40
		---
	*/
	buf.WriteString("---\n")
	buf.WriteString("note: GENERATED. DO NOT EDIT.\n")
	if cmd.Parent() == nil {
		buf.WriteString("title: CLI usage\n")
		buf.WriteString("sidebar_position: 15\n")
	} else {
		buf.WriteString("title: " + name + "\n")
	}
	buf.WriteString("---\n")

	buf.WriteString("## " + name + "\n\n")
	buf.WriteString(cmd.Short + "\n\n")
	if cmd.Long != "" {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.UseLine())
	}

	if cmd.Example != "" {
		buf.WriteString("### Examples\n\n")
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.Example)
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			var link string
			if parent.Parent() == nil {
				link = "cli.md"
			} else {
				link = parent.Name() + ".md"
			}

			if len(cmd.Commands()) > 0 {
				link = filepath.Join("..", link)
			}

			fmt.Fprintf(buf, "* [%s](%s)\t - %s\n", pname, link, parent.Short)
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}

			cname := name + " " + child.Name()
			link := child.Name() + ".md"
			if len(child.Commands()) > 0 {
				link = filepath.Join(child.Name(), link)
			}

			fmt.Fprintf(buf, "* [%s](%s)\t - %s\n", cname, link, child.Short)
		}
		buf.WriteString("\n")
	}

	_, err := buf.WriteTo(w)
	return err
}

func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Flags\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Global flags\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
