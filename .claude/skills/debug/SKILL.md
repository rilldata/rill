# Debug Skill

Investigate bugs by spawning parallel task agents to explore multiple hypotheses simultaneously.

## Usage

```
/debug <problem-description>
```

The problem description can be:
- **Free text**: Describe the bug directly
- **Slack link**: URL to a Slack message or thread
- **Linear issue**: Issue ID (e.g., `ENG-1234`) or full URL
- **Any combination**: Link + additional context or hypotheses

**Examples:**
```
/debug Getting 500 errors on the /api/metrics endpoint after deploying yesterday
```
```
/debug https://rill-data.slack.com/archives/C123/p1234567890
```
```
/debug ENG-1234
```
```
/debug ENG-1234 - I suspect it's related to the recent DuckDB upgrade, might be a connection pool issue
```
```
/debug https://rill-data.slack.com/archives/C123/p1234567890
The user mentions it only happens with large datasets. Could be memory or query timeout.
```

## Instructions

When this skill is invoked:

### 1. Parse Input and Gather Context

The input may contain multiple parts. Parse it to identify:

**Links to fetch:**
- **Slack link** (contains `slack.com/archives`): Use `mcp__slack__conversations_search_messages` with the URL to get the message, then `mcp__slack__conversations_replies` if it's a thread
- **Linear issue ID** (matches pattern like `ENG-1234`, `PROD-567`): Use `mcp__linear__get_issue` with the ID
- **Linear URL** (contains `linear.app`): Extract the issue ID from the URL and use `mcp__linear__get_issue`

**Additional context provided by the user:**
- Any text that isn't a link or issue ID is additional context
- This might include: suspected causes, recent changes, reproduction steps, or specific hypotheses to investigate
- Incorporate this context when generating hypotheses—user-provided hunches should be prioritized

**If the input is purely free text (no links/IDs):**
- Use the description directly as the problem statement
- Ask clarifying questions only if the description is too vague to form hypotheses

**Also fetch linked resources when applicable:**
- If a Linear issue has attachments, use `mcp__linear__get_attachment` to view them
- If there are screenshots/images in the description, use `mcp__linear__extract_images` to view them
- Check `mcp__linear__list_comments` for additional context from the team

### 2. Analyze and Generate Hypotheses

Based on the problem description, identify:
- The exact error or unexpected behavior
- Any error messages, stack traces, or logs mentioned
- The affected component/area of the codebase
- Recent changes that might be related

Generate 3-5 distinct hypotheses about the root cause:

**If the user provided specific hypotheses or hunches:**
- Include these as hypotheses (they have domain knowledge you don't)
- Add complementary hypotheses that cover other possibilities

**If starting from scratch:**
- Generate hypotheses that cover different categories: data issues, configuration, race conditions, resource limits, recent changes, etc.

Each hypothesis should be:
- Specific and testable
- Independent from other hypotheses
- Focused on a different potential cause

### 3. Spawn Parallel Investigation Agents

Use the **Task tool** to spawn multiple agents simultaneously (in a single message with multiple tool calls). Each agent should investigate one hypothesis.

For each agent:
- Use `subagent_type: "Explore"` for codebase investigation
- Give a clear, focused prompt describing:
  - The specific hypothesis to investigate
  - What files/code to examine
  - What evidence would confirm or refute the hypothesis
- Set `model: "sonnet"` for efficiency

**Example parallel spawn pattern:**
```
Task 1: "Investigate hypothesis: Database connection timeout"
  - Search for connection pool settings
  - Check timeout configurations
  - Look for recent changes to DB layer

Task 2: "Investigate hypothesis: Race condition in async handler"
  - Trace the request flow
  - Look for shared state mutations
  - Check locking mechanisms

Task 3: "Investigate hypothesis: Missing null check"
  - Find where the error originates
  - Trace data flow to find where null could enter
  - Check validation logic
```

### 4. Synthesize Findings

After all agents complete:
1. Summarize what each agent found
2. Identify which hypotheses were confirmed, refuted, or inconclusive
3. Recommend next steps:
   - If root cause found: Propose a fix
   - If inconclusive: Suggest additional investigation or reproduction steps
   - If multiple causes: Prioritize by likelihood

### 5. Output Format

Present findings in this structure:

```
## Problem Summary
[One paragraph describing the issue]

## Hypotheses Investigated

### Hypothesis 1: [Name]
**Status:** Confirmed / Refuted / Inconclusive
**Evidence:** [What the agent found]
**Files examined:** [List of relevant files with line numbers]

### Hypothesis 2: [Name]
...

## Conclusion
[Most likely root cause and reasoning]

## Recommended Action
[Specific next steps - either a fix or further investigation]
```

## Constraints

- Do NOT make code changes during investigation—this is read-only exploration
- Do NOT rename variables, refactor code, or make "improvements"
- If agents find the same issue, consolidate findings rather than duplicating
- If the problem description is unclear, ask clarifying questions before spawning agents
