---
note: GENERATED. DO NOT EDIT.
title: AI Eval YAML
sidebar_position: 42
---

AI evals define business questions with expected answers that test the project's AI assistant.
Create a `<eval_name>.yaml` file with `type: eval` (files under an `evals/` directory are inferred automatically).
Evals never run automatically; trigger them manually from Rill Developer to get pass/fail results per case.
Use them to verify that changes to `ai_instructions` fix problems without regressing answers that already work.


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `eval` _(required)_

### `display_name`

_[string]_ - Display name shown in the UI

### `notes`

_[string]_ - Suite-level guidance for the LLM judge, applied to every case

### `defaults`

_[object]_ - Defaults applied to every case

  - **`agent`** - _[string]_ - Agent that answers the questions. Currently only `analyst_agent`.

  - **`explore`** - _[string]_ - Explore dashboard used as context for all cases

### `concurrency`

_[integer]_ - Number of cases to run concurrently (default 2)

### `timeout`

_[string]_ - Timeout for a whole run (default 30m)

### `case_timeout`

_[string]_ - Timeout for a single case (default 5m)

### `cases`

_[array of object]_ - The test cases in this eval suite _(required)_

  - **`name`** - _[string]_ - Unique name of the case within this file _(required)_

  - **`question`** - _[string]_ - The question to ask the AI _(required)_

  - **`notes`** - _[string]_ - Case-level guidance for the LLM judge

  - **`explore`** - _[string]_ - Explore dashboard used as context for this case

  - **`expect`** - _[object]_ - Expected outcomes; every present key is asserted _(required)_

    - **`answer`** - _[string]_ - Expected answer in natural language, graded by an LLM judge

    - **`metrics_view`** - _[string]_ - Metrics view the AI must query

    - **`measures`** - _[array of string]_ - Measures the AI's queries must include

    - **`dimensions`** - _[array of string]_ - Dimensions the AI's queries must include

## Examples

```yaml
# Example: Revenue questions with judge and structural expectations
type: eval
display_name: Revenue questions
notes: Revenue always means net_revenue, never gross_bookings.
cases:
    - name: monthly_revenue_2024
      question: What was our total revenue per month in 2024?
      expect:
        answer: Monthly totals of net_revenue for all 12 months of 2024.
        metrics_view: sales
        measures:
            - net_revenue
```
