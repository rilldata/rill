// Command agent-demo demonstrates the ProjectEditor Agent with handoff to SyntheticDataAgent using OpenAI.
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
        <h1>🏗️ Rill Project Editor</h1>
        <p>Describe what you want to do with your project models:</p>
        <textarea id="prompt" placeholder="Examples:
• Generate synthetic sales data with customer info
• Edit the customer model to add demographics  
• Create a marketing campaign tracking model
• List all existing model files
• Generate sample e-commerce transaction data"></textarea>
        <div>
            <button onclick="generateSQL()" id="generateBtn">Process Request</button>
            <span id="loading" class="loading">Generating... (this may take a moment)</span>
        </div>
        <div>
            <h3>Result:</h3>
            <pre id="output">Your result will appear here...</pre>
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
                    throw new Error('Failed to process request');
                }

                const data = await response.json();
                output.textContent = data.sql || 'No result generated';
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
	model := flag.String("model", "gpt-4o", "OpenAI model to use")
	port := flag.String("port", "8080", "Port to run the web server on")
	projectDir := flag.String("project-dir", "./project", "Project directory for storing model files")
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

	// Create the ProjectEditor Agent with handoff capability
	projectEditor := agent.NewProjectEditorAgent(*model, runnerInstance, syntheticAgent, *projectDir)

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

		// Use the ProjectEditor agent to process the request with context
		log.Printf("Running ProjectEditor agent with prompt: %s", req.Prompt)

		ctx := context.Background()
		result, err := projectEditor.ProcessUserRequestWithContext(ctx, req.Prompt)
		if err != nil {
			errMsg := fmt.Sprintf("Error processing request: %v", err)
			log.Printf("%s", errMsg)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}

		log.Printf("ProjectEditor result: %+v", result)

		// Extract meaningful output from the result for the web interface
		var sqlOutput string

		// Handle different result types from ProjectEditor
		if resultMap, ok := result.(map[string]interface{}); ok {
			// Check for different handoff result types
			if handoffSuccessful, ok := resultMap["handoff_successful"].(bool); ok && handoffSuccessful {
				// Handle handoff results (ModelEditor or SyntheticData)
				if agentUsed, ok := resultMap["agent_used"].(string); ok {
					sqlOutput = fmt.Sprintf("Successfully handed off to %s\n\n", agentUsed)
				}

				// Add validation information for ModelEditor results
				if agentUsed, ok := resultMap["agent_used"].(string); ok && agentUsed == "ModelEditorAgent" {
					if validationPassed, ok := resultMap["validation_passed"].(bool); ok {
						if validationPassed {
							sqlOutput += "✅ Validation: PASSED\n"
						} else {
							sqlOutput += "❌ Validation: FAILED\n"
						}
					}
					
					if fileSaved, ok := resultMap["file_saved"].(bool); ok {
						if fileSaved {
							sqlOutput += "💾 File: SAVED\n"
						} else {
							sqlOutput += "⚠️  File: NOT SAVED (validation failed)\n"
						}
					}
					sqlOutput += "\n"
				}

				// Add the actual SQL/result content
				if updatedSQL, ok := resultMap["updated_sql"].(string); ok {
					sqlOutput += fmt.Sprintf("Updated SQL:\n%s", updatedSQL)
				} else if generatedSQL, ok := resultMap["generated_sql"].(string); ok {
					sqlOutput += fmt.Sprintf("Generated SQL:\n%s", generatedSQL)
				} else {
					sqlOutput += "Operation completed successfully"
				}

				// Add model information
				if modelName, ok := resultMap["model_name"].(string); ok {
					sqlOutput = fmt.Sprintf("Model: %s\n%s", modelName, sqlOutput)
				}
			} else {
				// Handle other map results
				sqlOutput = fmt.Sprintf("Result: %+v", resultMap)
			}
		} else if str, ok := result.(string); ok {
			// Handle direct string results
			sqlOutput = str
		} else {
			// Handle any other result type
			sqlOutput = fmt.Sprintf("Result: %+v", result)
		}

		// Return the result as JSON
		json.NewEncoder(w).Encode(GenerateResponse{
			SQL: sqlOutput,
		})
	})

	// Start the server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("🚀 Server running on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
