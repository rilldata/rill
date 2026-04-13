# Cagan-Informed Product Management Skill Stack

## Purpose

Build a Marty Cagan-informed Product Management skill stack that mirrors the existing design skill stack's adversarial debate model. Each skill is a separate voice/stakeholder that debates the best path forward. Two leadership synthesis skills (CPO and CDO) weigh the debates and make recommendations.

## Decisions

- **Depth model:** Option A — Full reference directory for the backbone skill (`product-thinking`), with 7 reference files providing deep, citable theory. Other skills are self-contained SKILL.md files.
- **Relationship to existing skills:** Option A — Wrapper model. New skills form a philosophical layer around existing PM utilities (`prd-generator`, `stakeholder-review`, `feature-prioritization`, `okr-writer`, `user-story`, `go-to-market`, `metrics-framework`, `feedback-analysis`, `competitive-analysis`, `executive-update`). They gate, challenge, and validate but never replace.
- **Source fidelity:** Option B — Cagan-inspired, practitioner-oriented. Uses Cagan's core ideas as foundation but writes references as a practitioner would apply them: opinionated, concrete, with decision trees and anti-pattern tables.

## Architecture

### File Structure

```
~/.claude/skills/
  product-thinking/          # Philosophy backbone (like design-for-ai)
    SKILL.md
    references/
      four-risks.md
      discovery-vs-delivery.md
      empowered-teams.md
      opportunity-assessment.md
      anti-patterns.md
      evidence-levels.md
      decision-trees.md
  product-discovery/         # Gatekeeper: "should we build this?"
    SKILL.md
  product-strategy/          # Vision -> strategy -> goals
    SKILL.md
  product-review/            # 4-risk validation audit
    SKILL.md
  product-devil/             # Adversarial skeptic
    SKILL.md
  product-pulse/             # Post-launch continuous learning
    SKILL.md
  product-lead/              # CPO synthesis voice
    SKILL.md
  design-lead/               # CDO synthesis voice
    SKILL.md

~/.claude/commands/
  product-thinking.md
  product-discovery.md
  product-strategy.md
  product-review.md
  product-devil.md
  product-pulse.md
  product-lead.md
  design-lead.md
```

### Debate Topology

```
                     product-thinking
                    (philosophy backbone)
                   /    |    |    |    \
                  /     |    |    |     \
     product-     product-  product-  product-  product-
     discovery    strategy  review    devil     pulse
         |            |        |        |         |
         v            v        v        v         v
    prd-generator  okr-writer  stakeholder-  feature-    metrics-
    user-story     go-to-market  review     prioritization framework
                                                         feedback-analysis

                        product-lead
                   (CPO: synthesizes PM debate,
                    makes the call)

                        design-lead
                   (CDO: synthesizes design debate,
                    makes the call)
```

### Tension Lines

1. **Discovery vs Execution:** `product-discovery` blocks `prd-generator` ("You haven't validated the problem"). `prd-generator` pushes back ("We have enough signal to spec").
2. **Strategy vs Skepticism:** `product-strategy` says "This connects to our vision." `product-devil` says "Your vision-to-feature logic has three leaps of faith."
3. **Review vs Stakeholders:** `product-review` applies the 4-risk framework with depth. `stakeholder-review` applies 7 perspectives with breadth. They find different problems.
4. **Pulse vs Metrics:** `product-pulse` asks "Are we learning what we need to learn?" `metrics-framework` asks "Are the numbers trending right?" These are different questions.
5. **Discovery vs Strategy:** `product-discovery` works bottom-up from user problems. `product-strategy` works top-down from company vision. The best product decisions survive both.

## Skill Specifications

### Skill 1: `product-thinking` (Philosophy Backbone)

**Role:** The Cagan equivalent of `design-for-ai`. Every other PM skill references this. Two modes: CHECKER (audit a product decision) and ADVISOR (guide a new one).

**CHECKER mode phases:**
1. Discovery hygiene: was the problem validated before solutioning?
2. Risk coverage: are all four risks (value, usability, feasibility, viability) addressed?
3. Team dynamics: is this an empowered team or a feature team taking orders?
4. Evidence quality: what level of evidence supports the key assumptions?
5. Outcome orientation: is success defined as an outcome or an output?

**ADVISOR mode phases:**
1. Frame the opportunity (not the solution)
2. Identify the riskiest assumption
3. Design the cheapest test for that assumption
4. Route to the right skill for next steps

**Reference files (7):**

- `four-risks.md`: Value risk ("Will customers buy/use it?"), usability risk ("Can they figure it out?"), feasibility risk ("Can we build it?"), viability risk ("Does it work for the business?"). For each: how to identify, how to test, common failure modes, evidence thresholds.

- `discovery-vs-delivery.md`: What discovery actually is (reducing risk before committing resources) vs what teams think it is (having a meeting about requirements). Decision tree: "Have you done discovery?" with concrete checkpoints. Anti-patterns: "We did discovery; we asked the stakeholder what they want."

- `empowered-teams.md`: Empowered teams own problems, not features. Feature teams get handed solutions. Diagnostic: which one are you? Signs of a feature team pretending to be empowered. What to do when you're on a feature team.

- `opportunity-assessment.md`: Before any work: who is it for, what problem does it solve, how will we know it worked, what's the opportunity cost? Template with forced constraints (each answer <= 2 sentences).

- `anti-patterns.md`: The 10 most common ways teams fake good product process: roadmap-driven development, stakeholder-driven priorities, consensus-as-strategy, metrics theater, solution-first thinking, discovery theater, PRD-as-discovery, the "just build it" shortcut, output celebration, and the feature factory.

- `evidence-levels.md`: Hierarchy from weakest to strongest: opinion -> anecdote -> survey -> behavioral data -> prototype test -> live experiment -> sustained metric movement. What each level is sufficient for.

- `decision-trees.md`: Routing logic: given where you are in the product process, which skill to invoke next. Maps common starting points to the right skill sequence.

### Skill 2: `product-discovery` (Gatekeeper)

**Role:** The "should we build this at all?" gate. Runs before `prd-generator`. Forces evidence-based justification.

**Process:**
1. Frame the opportunity: Who has the problem? How do we know? What's the current behavior?
2. Map assumptions: List every assumption baked into the proposal. Rank by risk x ignorance.
3. Demand evidence: For each critical assumption: what evidence exists? What level? (References `evidence-levels.md`)
4. Identify cheapest test: What's the minimum effort to de-risk the top assumption?
5. Gate verdict: one of:
   - **Proceed to PRD**: enough evidence, risks understood -> route to `prd-generator`
   - **Test first**: critical assumptions untested -> prescribe specific experiments
   - **Kill**: evidence actively contradicts the premise -> explain why, suggest alternatives

**Anti-rationalization table:**

| Excuse | Reality |
|--------|---------|
| "The customer asked for it" | Customers describe symptoms, not solutions. What job are they hiring this feature to do? |
| "Leadership wants it" | Leadership sets problems and constraints, not solutions. What outcome are they after? |
| "Competitors have it" | Competitors may be wrong. They may also have different users, strategy, or context. |
| "It's obvious" | If it's obvious, the evidence should be trivial to produce. Produce it. |
| "We don't have time to test" | You don't have time to build the wrong thing. Testing is faster than building. |
| "We'll learn by shipping" | Shipping is the most expensive way to learn. What can you learn before shipping? |

### Skill 3: `product-strategy` (Vision -> Strategy -> Goals)

**Role:** Ensures features connect to strategy. Works top-down: company mission -> product vision -> product strategy -> objectives -> key results -> features.

**Process:**
1. State the chain: For any proposed work, articulate the full chain from mission to feature. Flag missing or weak links.
2. Strategy test: A strategy is a set of bets about how to win. Can you state the bets? What would prove them wrong?
3. Opportunity cost: What are you NOT doing by doing this? Is what you're giving up worth less than what you're gaining?
4. Time horizon check: Now-bet (evidence-backed, near-term) vs future-bet (conviction-backed, long-term). Both valid; conflating them is not.
5. Route: If the strategy link is solid, route to `okr-writer`. If not, route back to `product-discovery`.

**Debates with:**
- `product-discovery`: Discovery says "users need X." Strategy says "X doesn't connect to our bets." Both can be right.
- `product-devil`: Strategy says "this is our bet." Devil says "your bet is based on three unvalidated assumptions."

### Skill 4: `product-review` (4-Risk Validation Audit)

**Role:** After a PRD or proposal exists, audit it through the four-risk lens. Theory-backed, like `/exam` for design.

**For each risk, evaluate:**

**Value risk:** Does anyone actually want this?
- What evidence exists that users have this problem?
- What's the current workaround? How painful is it?
- Have we tested demand (not just collected opinions)?
- Severity: Critical (no evidence) / Major (weak evidence) / Clear (strong evidence)

**Usability risk:** Can users figure it out?
- Has the interaction model been prototyped and tested?
- Are we introducing new concepts users must learn?
- What's the failure mode if a user gets confused?

**Feasibility risk:** Can we actually build this?
- Has engineering spiked the hard parts?
- Are there dependencies on systems we don't control?
- What's the confidence interval on the timeline?

**Viability risk:** Should we build this?
- Does it work with our business model?
- Legal, compliance, ethical concerns?
- Does it create maintenance burden disproportionate to value?

**Output:** Risk scorecard with findings table organized by severity, then suggested next steps.

### Skill 5: `product-devil` (Adversarial Skeptic)

**Role:** The designated contrarian. Finds the strongest case AGAINST the current direction.

**Process:**
1. Steel-man the proposal: Articulate the best version of the argument FOR building this.
2. Attack the assumptions: For each key assumption, construct the case that it's wrong.
3. Inversion test: "What would have to be true for this to fail?" List failure conditions.
4. Opportunity cost attack: "What's the best alternative use of this team's time?" Make the case.
5. Pre-mortem: "It's 6 months from now and this failed. Write the post-mortem."
6. Verdict: one of:
   - **Conviction holds**: the case against is weak; proceed with confidence
   - **Hedgeable**: real risks exist but can be mitigated; here's how
   - **Reconsider**: the case against is stronger than the case for; here's why

**Fights these biases:** confirmation bias, sunk cost, authority bias, groupthink, narrative fallacy.

### Skill 6: `product-pulse` (Post-Launch Continuous Learning)

**Role:** After shipping, close the loop. Connects back to discovery hypotheses.

**Process:**
1. Recall the hypothesis: What did we believe would happen?
2. Check reality: What actually happened?
3. Diagnosis: Validated / Partially validated / Invalidated / Inconclusive
4. Learning extraction: What changes about our understanding?
5. Next action: Route to `metrics-framework`, `product-discovery`, or `product-strategy`.

**Debates with:**
- `metrics-framework`: Pulse asks "what did we learn?" Metrics asks "what did we measure?"
- `product-strategy`: Pulse may surface evidence that invalidates a strategic bet.

### Skill 7: `product-lead` (CPO Synthesis Voice)

**Role:** Runs after any PM skill debate. Weighs competing perspectives, makes a clear recommendation, owns the tradeoff.

**How it thinks:**
1. Acknowledge the tensions: doesn't pretend disagreement doesn't exist
2. Weigh by context: early-stage (discovery risk dominates), mature (viability risk dominates), turnaround (strategy trumps incrementalism)
3. Make the call: "Here's what I'd do and why." Not consensus. A decision.
4. State the reversibility: one-way door (high stakes, get it right) vs two-way door (ship it, learn, adjust)
5. Prescribe next steps: exactly which skill/command to run next, in what order

**Voice:** Experienced, decisive, comfortable with ambiguity. Doesn't hedge. Says "I'd ship this despite the discovery gap because the opportunity cost of waiting exceeds the risk" and explains why.

### Skill 8: `design-lead` (CDO Synthesis Voice)

**Role:** Same pattern for the design stack. After design skills surface tensions, synthesizes and recommends.

**How it thinks:**
1. Weigh aesthetic vs functional: when `brand` wants character but `design-qa` flags accessibility, which wins and why?
2. Context-aware prioritization: internal tool (ship fast, nail usability), consumer product (invest in delight), enterprise (consistency and trust)
3. Make the call: "The brand concern is valid but secondary here. Nail the interaction states first, add character in the next pass."
4. State what's good enough: knows when to stop polishing. "This is shippable. These three things can improve post-launch."
5. Prescribe next steps: which design command to run next

## Debate Model in Practice

When any PM skill is invoked, the output includes:

1. **The skill's own analysis and verdict**
2. **Where other skills would disagree** (2-3 tension points, not all of them)
3. **`product-lead` recommendation** ("Here's the call I'd make and why")
4. **Prescribed next step** (which command to run next)

Example:

```
You: /product-discovery "We should add a Slack integration"

product-discovery runs its gatekeeper process...

=== Discovery Verdict: TEST FIRST ===
No evidence users want this over email. Two assumptions untested:
1. Users check Slack more than email for alerts
2. Alert response time matters for the use case

Cheapest test: instrument email alert open rates, survey 20 users on preferred channel.

=== Debate ===
product-strategy would say: "Slack integration connects to our 'meet users where
they are' bet. Strategic fit is strong."

product-devil would say: "Three competitors added Slack integrations last year.
Two removed them within 6 months. Why?"

=== product-lead recommendation ===
I'd run the test. Strategic fit is real, but the devil's point about competitor
removals is a red flag worth 1 week of investigation. This is a two-way door;
we can ship a lightweight version later if the signal is strong. But building
a full integration without evidence is a 6-week bet on an assumption.

Next: survey 20 power users on alert channel preference. If >60% say Slack,
run /prd-generator. If not, revisit the problem framing.
```

Same pattern for design:

```
After design skills debate...

=== design-lead recommendation ===
Ship the current color system. The brand concern about generic palette is real
but secondary to the interaction state gaps that exam flagged. Fix hover/focus/
disabled states first (run /flow), then revisit palette character in the next
design pass (run /brand).
```

## Implementation Order

1. `product-thinking` (backbone + all 7 reference files) — everything else depends on this
2. `product-lead` + `design-lead` (synthesis layer) — needed by all other skills for output format
3. `product-discovery` (gatekeeper) — highest-impact individual skill
4. `product-devil` (skeptic) — second-highest impact, enables debate quality
5. `product-review` (4-risk audit) — validates existing PRDs/proposals
6. `product-strategy` (vision-to-feature chain) — connects to existing `okr-writer`
7. `product-pulse` (post-launch) — closes the learning loop
8. Slash commands (all 8) — thin wrappers, built last

## Success Criteria

- Each skill produces output that surfaces disagreement from at least 2 other skills
- `product-lead` and `design-lead` make decisive recommendations, not wishy-washy summaries
- The debate model feels like a leadership team arguing, not a committee reaching consensus
- Existing PM skills (`prd-generator`, `stakeholder-review`, etc.) continue to work unchanged
- A user can invoke any single skill and get value; the full debate topology is optional depth
