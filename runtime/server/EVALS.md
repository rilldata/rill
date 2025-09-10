# AI Chat Evals

This directory contains evaluation tests for Rill's AI chat completion functionality, specifically testing the project-level prompt behavior.

## Overview

The eval suite tests that the AI assistant follows the systematic analysis process defined in the project prompt and provides high-quality, actionable insights.

## Running Evals

### Prerequisites

1. Set your OpenAI API key (uses Rill's standard environment variable):
   ```bash
   export RILL_ADMIN_OPENAI_API_KEY="your-api-key-here"
   # Or alternatively:
   export OPENAI_API_KEY="your-api-key-here"
   ```

2. Optionally set the model to use (defaults to `gpt-4o`):
   ```bash
   export EVAL_MODEL="gpt-4o"  # or gpt-4o-mini, gpt-4-turbo, etc.
   ```

### Local Testing

Run the eval suite locally:

```bash
# Run all evals (recommended)
go test -tags=evals ./runtime/server/ -run TestProjectPromptEvals -v -timeout=30m

# Run specific eval test
go test -tags=evals ./runtime/server/ -run TestProjectPromptEvals/follows_systematic_analysis_process -v -timeout=30m

# Alternative: target just the eval test file
go test -tags=evals ./runtime/server/chat_eval_test.go -v -timeout=30m
```

### CI Integration

The evals run automatically on:
- Manual trigger via GitHub Actions workflow dispatch
- PRs that modify `runtime/completion.go` or eval-related files

To manually trigger evals in CI:
1. Go to the Actions tab in GitHub
2. Select "AI Evals" workflow  
3. Click "Run workflow"
4. Choose the OpenAI model and run

## Test Structure

### Core Eval Tests

1. **`follows_systematic_analysis_process`** - Verifies the AI follows the 4-phase process:
   - Discovery: Uses `list_metrics_views`
   - Understanding: Uses `get_metrics_view`
   - Scoping: Uses `query_metrics_view_time_range`
   - Analysis: Performs multiple `query_metrics_view` calls

2. **`provides_quantified_actionable_insights`** - Uses LLM-as-a-judge to evaluate insight quality

3. **`uses_comparison_features_for_time_analysis`** - Ensures time-based analysis uses comparison features

4. **`maintains_analytical_rigor`** - Verifies the AI only uses numbers from tool results

5. **`formats_output_correctly`** - Checks adherence to the markdown output format

### LLM-as-a-Judge

The eval suite uses "LLM-as-a-judge" methodology where the same OpenAI model evaluates response quality against specific criteria. Responses are scored 1-10 with a passing threshold of 7.0.

## Test Data

The evals use the **OpenRTB programmatic advertising project** located in `runtime/testruntime/testdata/openrtb/` with:
- Auction and bid stream data (realistic programmatic advertising dataset)
- Dimensions: advertiser, domain, device_type, country
- Metrics: total_bids, total_wins, total_spend, avg_bid_price, win_rate
- Time range: One week of sampled data
- Complex business context perfect for testing analytical reasoning

This is located alongside other runtime test projects and can be modified for backend eval needs without affecting frontend tests.

## Adding New Evals

To add new evaluation tests:

1. Add a new test function in `runtime/server/chat_eval_test.go`
2. Use the `runProjectChatCompletion()` helper for consistency
3. Consider using `evaluateWithLLM()` for subjective quality assessment
4. Update this documentation

## Cost Considerations

- Each eval test makes multiple OpenAI API calls
- Estimated cost: $0.10-0.50 per full test run (depending on model and response length)
- Tests are excluded from regular CI runs to control costs
- Use `gpt-4o-mini` for faster/cheaper development iteration

## Troubleshooting

**Build tag warnings**: The `//go:build evals` tag means IDEs may not recognize the file. This is expected behavior.

**API key issues**: Ensure `RILL_ADMIN_OPENAI_API_KEY` or `OPENAI_API_KEY` is set and valid. The tests will skip if neither is provided.

**Timeout errors**: Increase the timeout with `-timeout=45m` for comprehensive evals.

**Real MCP Integration**: Uses the actual MCP server to provide authentic tool responses from the OpenRTB metrics views, giving realistic eval results.
