package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ChatCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var local bool

	chatCmd := &cobra.Command{
		Use:               "chat [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Chat with the Rill AI",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project name
			if len(args) > 0 {
				project = args[0]
			}
			if !local && project == "" {
				if !ch.Interactive {
					return fmt.Errorf("project not specified and could not be inferred from context")
				}
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return err
			}

			// Change to a ctx that needs two ctrl+C in succession to trigger.
			// This allows users to cancel streaming responses without exiting the chat.
			ctx, cancel := newDoubleInterruptContext(context.Background())
			defer cancel()

			// Main chat loop
			printWelcome()
			scanner := bufio.NewScanner(os.Stdin)
			var conversationID string
			var prevCancel context.CancelFunc
			for {
				// If the parent context is cancelled, we quit
				if ctx.Err() != nil {
					break
				}

				// If we receive a single ctrl+C, we want to stop the current loop and continue to the next loop.
				// We only truly quit if we get two ctrl+C in a row. See use of newDoubleInterruptContext above.
				ctx, cancel := newInterruptContext(ctx)
				if prevCancel != nil {
					prevCancel()
				}
				prevCancel = cancel

				// Show prompt and read input
				fmt.Printf("\n> ")
				input, err := scanContext(ctx, scanner)
				if err != nil {
					if errors.Is(err, ctx.Err()) {
						continue
					}
					if errors.Is(err, io.EOF) {
						break
					}
					return fmt.Errorf("failed to read input: %w", err)
				}
				if input == "" {
					continue
				}

				// Call the completion
				stream, err := rt.CompleteStreaming(ctx, &runtimev1.CompleteStreamingRequest{
					InstanceId:     instanceID,
					ConversationId: conversationID,
					Prompt:         input,
				})
				if err != nil {
					if errors.Is(err, ctx.Err()) {
						continue
					}
					return fmt.Errorf("completion failed: %w", err)
				}

				// Handle streaming messages
				var lastMsg *runtimev1.Message
				for {
					// Handle end of stream
					resp, err := stream.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						if errors.Is(err, ctx.Err()) {
							break
						}
						if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
							break
						}
						return fmt.Errorf("completion stream failed: %w", err)
					}

					// Print truncated message
					typ, text := formatMessage(resp.Message)
					n := terminalWidth() - 4 - len(typ)
					text = truncateString(text, n)
					fmt.Printf("\n◆ %s: %s\n", typ, text)

					// Update state
					conversationID = resp.ConversationId // Won't change after first iteration
					lastMsg = resp.Message
				}

				// Print the last message untruncated
				if lastMsg != nil {
					_, text := formatMessage(lastMsg)
					fmt.Print("\033[3A\033[0J") // Clear the previous three lines
					fmt.Printf("◆ %s\n", text)
				}
			}

			fmt.Printf("\nGoodbye!\n")
			return nil
		},
	}

	chatCmd.Flags().SortFlags = false
	chatCmd.Flags().StringVar(&project, "project", "", "Project name")
	chatCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	chatCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")

	return chatCmd
}

// printWelcome prints a welcome message with usage tips.
func printWelcome() {
	fmt.Println("╭─────────────────────────────────────────────╮")
	fmt.Println("│ Welcome to Rill AI Chat                     │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│ Tips:                                       │")
	fmt.Println("│ • Type your message and press Enter.        │")
	fmt.Println("│ • Use ctrl+C to cancel a response.          │")
	fmt.Println("│ • Enter ctrl+C twice to quit.               │")
	fmt.Println("╰─────────────────────────────────────────────╯")
}

// formatMessage formats a runtimev1.Message for display.
func formatMessage(msg *runtimev1.Message) (string, string) {
	if msg == nil || len(msg.Content) == 0 {
		return "empty", ""
	}

	typ := "assistant"
	content := msg.ContentData

	if msg.ContentType == string(ai.MessageContentTypeError) {
		content = fmt.Sprintf("Error: %s", msg.ContentData)
	}

	if msg.Type == string(ai.MessageTypeCall) {
		typ = fmt.Sprintf("call(%s)", msg.Tool)
	} else if msg.Type == string(ai.MessageTypeResult) {
		typ = fmt.Sprintf("result(%s)", msg.Tool)
	}

	if msg.Tool == string(ai.RouterAgentName) {
		if msg.Type == string(ai.MessageTypeResult) {
			typ = "response"
			if msg.ContentType == string(ai.MessageContentTypeJSON) {
				var response struct {
					Response string `json:"response"`
				}
				if err := json.Unmarshal([]byte(msg.ContentData), &response); err == nil {
					content = response.Response
				}
			}
		}
	}

	return typ, content
}

// truncateString truncates a string to a maximum width, adding ellipsis if needed.
func truncateString(msg string, maxWidth int) string {
	msg = strings.ReplaceAll(msg, "\n", "\\n")
	msg = strings.Join(strings.Fields(msg), " ")
	if len(msg) <= maxWidth {
		return msg
	}
	return msg[:maxWidth-3] + "..."
}

// terminalWidth returns the current width of the terminal.
func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// scanContext scans user input with support for context cancellation.
func scanContext(ctx context.Context, scanner *bufio.Scanner) (string, error) {
	scanCh := make(chan bool)
	go func() {
		scanCh <- scanner.Scan()
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case ok := <-scanCh:
		if ok {
			return strings.TrimSpace(scanner.Text()), nil
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
}

// newInterruptContext returns a context that is canceled when an interrupt signal is received.
func newInterruptContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt)
}

// newDoubleInterruptContext returns a context that requires two interrupt signals received within one second of each other to cancel.
func newDoubleInterruptContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		defer signal.Stop(sigChan)

		var lastInterrupt time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigChan:
				now := time.Now()
				if now.Sub(lastInterrupt) > time.Second {
					lastInterrupt = now
					continue
				}
				cancel()
				return
			}
		}
	}()

	return ctx, cancel
}
