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
	"strings"

	"github.com/spf13/cobra"
)

func GenerateCmd(rootCmd *cobra.Command) *cobra.Command {
	docsCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate CLI documentation",
		Args:  cobra.ExactArgs(1),
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
	identity := func(s string) string {
		parts := strings.Split(s, "_")
		last := &parts[len(parts)-1]
		*last = (*last)[:len(*last)-3]
		return filepath.Join(parts...)
	}
	emptyStr := func(s string) string { return "" }
	return genMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

func genMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
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

		if _, err := os.Stat(sd); os.IsNotExist(err) {
			if err := os.Mkdir(sd, fs.ModePerm); err != nil {
				return err
			}
		}

		if err := genMarkdownTreeCustom(c, sd, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	nm := cmd.Name()
	filename := filepath.Join(dir, nm+".md")
	if len(cmd.Commands()) > 0 && cmd.Parent() != nil {
		filename = filepath.Join(dir, cmd.Name(), nm+".md")
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.WriteString(filePrepender(filename)); err != nil {
		return err
	}

	return genMarkdownCustom(cmd, f, linkHandler)
}

func genMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString("## " + name + "\n\n")
	buf.WriteString(cmd.Short + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := parent.Name() + ".md"
			if len(cmd.Commands()) > 0 {
				link = filepath.Join("..", link)
			}

			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, link, parent.Short))
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

			buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, link, child.Short))
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

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
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
