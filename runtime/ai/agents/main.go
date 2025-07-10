// Command agent-demo demonstrates the SyntheticDataAgent with OpenAI.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/sashabaranov/go-openai"

	"github.com/rilldata/rill/runtime/ai/agents/agent"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Synthetic Data Generator</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .container { display: flex; flex-direction: column; gap: 20px; }
        textarea { width: 100%; min-height: 100px; padding: 10px; font-family: monospace; }
        button { padding: 10px 15px; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:disabled { background: #cccccc; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 4px; overflow-x: auto; }
        .loading { display: none; color: #666; font-style: italic; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🦆 Synthetic Data Generator</h1>
        <p>Describe the data you want to generate in natural language:</p>
        <textarea id="prompt" placeholder="e.g., Create a table for e-commerce orders with order ID, customer details, and product information"></textarea>
        <div>
            <button onclick="generateSQL()" id="generateBtn">Generate SQL</button>
            <span id="loading" class="loading">Generating... (this may take a moment)</span>
        </div>
        <div>
            <h3>Generated SQL:</h3>
            <pre id="output">Your generated SQL will appear here...</pre>
        </div>
    </div>

    <script>
        async function generateSQL() {
            const prompt = document.getElementById('prompt').value.trim();
            if (!prompt) {
                alert('Please enter a prompt');
                return;
            }

            const output = document.getElementById('output');
            const btn = document.getElementById('generateBtn');
            const loading = document.getElementById('loading');

            btn.disabled = true;
            loading.style.display = 'inline';
            output.textContent = 'Generating...';

            try {
                const response = await fetch('/generate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ prompt })
                });

                if (!response.ok) {
                    throw new Error('Failed to generate SQL');
                }

                const data = await response.json();
                output.textContent = data.sql || 'No SQL generated';
            } catch (error) {
                output.textContent = 'Error: ' + error.message;
                console.error('Error:', error);
            } finally {
                btn.disabled = false;
                loading.style.display = 'none';
            }
        }

        // Allow Ctrl+Enter to submit
        document.getElementById('prompt').addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.key === 'Enter') {
                generateSQL();
            }
        });
    </script>
</body>
</html>
`

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

type GenerateResponse struct {
	SQL string `json:"sql"`
}

// openAIProvider implements the model.Provider interface
type openAIProvider struct {
	client *openai.Client
	model  string
}

// GetModel implements model.Provider
func (p *openAIProvider) GetModel(modelName string) (model.Model, error) {
	if modelName == "" {
		modelName = p.model
	}
	return &openAIModel{
		client: p.client,
		model:  modelName,
	}, nil
}

// openAIModel implements the model.Model interface
type openAIModel struct {
	client *openai.Client
	model  string
}

// GetResponse implements model.Model
func (m *openAIModel) GetResponse(ctx context.Context, req *model.Request) (*model.Response, error) {
	// Convert the request to OpenAI format
	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: req.SystemInstructions,
		},
	}

	// Add user input
	if input, ok := req.Input.(string); ok {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: input,
		})
	}

	// Call the OpenAI API
	request := openai.ChatCompletionRequest{
		Model:       m.model,
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	log.Printf("Making OpenAI API call with model: %s", m.model)
	resp, err := m.client.CreateChatCompletion(ctx, request)
	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return nil, fmt.Errorf("openai completion error: %w", err)
	}

	// Convert the response
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &model.Response{
		Content: resp.Choices[0].Message.Content,
	}, nil
}

// StreamResponse implements model.Model
func (m *openAIModel) StreamResponse(ctx context.Context, req *model.Request) (<-chan model.StreamEvent, error) {
	// For simplicity, we'll just call GetResponse and wrap it in a channel
	// In a real implementation, you'd want to use the streaming API
	resp, err := m.GetResponse(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan model.StreamEvent, 1)
	ch <- model.StreamEvent{
		Type:    "content",
		Content: resp.Content,
		Done:    true,
	}
	close(ch)
	return ch, nil
}

func main() {
	// Parse command line flags
	apiKey := flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
	model := flag.String("model", "gpt-4o-mini", "OpenAI model to use")
	port := flag.String("port", "8080", "Port to run the web server on")
	flag.Parse()

	// Get API key from flag or environment variable
	if *apiKey == "" {
		*apiKey = os.Getenv("OPENAI_API_KEY")
		if *apiKey == "" {
			log.Fatal("OpenAI API key is required. Set it via -api-key flag or OPENAI_API_KEY environment variable")
		}
	}

	// Create the OpenAI provider
	provider := &openAIProvider{
		client: openai.NewClient(*apiKey),
		model:  *model,
	}

	// Create a new runner with the provider
	runnerInstance := runner.NewRunner()
	runnerInstance.WithDefaultProvider(provider)

	// Create the SyntheticDataAgent with the model name and runner
	syntheticAgent := agent.NewSyntheticDataAgent(*model, runnerInstance)

	// HTTP Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(htmlTemplate))
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Use the agent's GenerateSQL method which includes validation
		log.Printf("Running agent with prompt: %s", req.Prompt)
		
		output, err := syntheticAgent.GenerateSQL(req.Prompt)
		if err != nil {
			errMsg := fmt.Sprintf("Error generating SQL: %v", err)
			log.Printf("%s", errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		log.Printf("Final validated SQL: %s", output)

		// Return the result as JSON
		json.NewEncoder(w).Encode(GenerateResponse{
			SQL: output,
		})
	})

	// Start the server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("🚀 Server running on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
