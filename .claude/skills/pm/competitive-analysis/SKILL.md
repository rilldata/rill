---
description: Conduct a competitive analysis using Gibson Biddle's DHM (Delight customers, Hard to copy, Margin-enhancing) framework, with strategic positioning recommendations
allowed-tools: Read, AskUserQuestion, WebSearch, Task
argument-hint: "<your product/company and list of competitors, or just the product space>"
---

Build a comprehensive competitive analysis using Gibson Biddle's DHM framework — researching up to 5 competitors in parallel and synthesizing strategic positioning recommendations.

Input: $ARGUMENTS

## Instructions

### 1. Define the Competitive Landscape

If competitors aren't specified, ask via `AskUserQuestion`:
- What product or space are we analyzing?
- Who are the top 3–5 competitors (direct and adjacent)?
- What is the primary job-to-be-done we're competing on?

### 2. Research Competitors in Parallel

Use the `Task` tool to spawn parallel `Explore` agents — one per competitor — each tasked with finding:
- Product positioning and key messaging
- Pricing model and tiers
- Key features and differentiators
- Known strengths and weaknesses (reviews, press, community feedback)
- Recent product moves (launches, pivots, funding)

Consolidate findings before proceeding.

### 3. DHM Analysis

For each competitor (including your own product), evaluate across the three DHM dimensions:

**Delight (D)**: What features or experiences genuinely delight customers? What do users love about this product?

**Hard to Copy (H)**: What aspects of their product, data network, brand, or business model are difficult for a new entrant to replicate? Rate each: Low / Medium / High defensibility.

**Margin-enhancing (M)**: How does this product structure itself to improve margins over time? (e.g., automation, self-serve, network effects, data flywheel)

### 4. Feature Comparison Matrix

Create a table comparing all competitors across key dimensions relevant to the space. Mark: ✅ (has it), ⚠️ (partial), ❌ (missing), 🔄 (in progress/announced).

### 5. Pricing Comparison

Document each competitor's pricing structure:
- Free tier (if any)
- Core paid plan
- Enterprise pricing signals
- Pricing model (seat-based, usage-based, flat, hybrid)

Note where your product is priced relative to the field and whether that positioning is strategic or legacy.

### 6. Strategic Recommendations

Synthesize the analysis into actionable recommendations:

**White spaces**: Jobs-to-be-done that no competitor is excelling at — potential areas to own.

**Defensive moves**: Where are competitors gaining ground that threatens your core? What should you protect?

**Positioning sharpening**: Based on the DHM analysis, what is the single most defensible and differentiating position for your product to own?

**Features to not build**: Capabilities where competitors have too much of a head start — better to partner, acquire, or integrate than build.

## Output Format

Structure as a competitive brief with: executive summary, DHM scorecards, feature matrix, pricing table, and strategic recommendations. Use tables extensively. Keep the executive summary to 5 bullets — the kind a CEO would read in 2 minutes.
