---
description: Generate sprint-ready user stories with acceptance criteria, story point estimates, and edge cases from a feature description or PRD section
allowed-tools: Read, AskUserQuestion
argument-hint: "<feature description, PRD section, or requirement to break into stories>"
---

Convert feature descriptions or PRD requirements into sprint-ready user stories with clear acceptance criteria, story point estimates, and edge cases documented — ready to paste into Linear or Jira.

Input: $ARGUMENTS

## Instructions

### 1. Understand the Feature

If the input is a file, use `Read`. If it's vague, ask via `AskUserQuestion`:

- Who is the primary user of this feature?
- What is the "happy path" — the most common way this will be used?
- Are there any known constraints (tech limitations, non-negotiable behaviors)?
- What does "done" look like from a user perspective?

### 2. Epic Summary (if applicable)

If the input describes a large feature, first write a one-paragraph **epic summary**:

- What is being built and why
- Who it's for
- What the scope boundaries are (what is NOT included)

### 3. Break Into Stories

Apply the INVEST criteria to each story:

- **Independent**: Can be developed without requiring another story to be done first
- **Negotiable**: Details can be discussed with engineering
- **Valuable**: Delivers something meaningful to the user or system
- **Estimable**: Can be pointed
- **Small**: Can be completed in one sprint (ideally 1–3 days of work)
- **Testable**: Has clear, verifiable acceptance criteria

Story format:

```
As a [user type],
I want to [perform an action],
so that [I achieve a benefit/outcome].
```

### 4. Acceptance Criteria

For each story, write 3–6 acceptance criteria using Given/When/Then:

```
Given [precondition or starting state]
When [the user takes an action]
Then [the expected result occurs]
```

Cover:

- The happy path
- Error states and validation
- Boundary conditions (empty states, max limits)
- Permission or role-based behavior (if relevant)

### 5. Story Point Estimates

Estimate each story using Fibonacci (1, 2, 3, 5, 8):

- **1 pt**: Trivial, under 2 hours, no unknowns
- **2 pt**: Small, half a day, minimal complexity
- **3 pt**: Medium, 1–2 days, some decisions to make
- **5 pt**: Large, 3–4 days, multiple components, some uncertainty
- **8 pt**: Very large — consider splitting before sprint planning

For anything 8+, flag it as a splitting candidate and suggest how to break it up.

### 6. Edge Cases and Notes

For each story, list:

- **Edge cases**: Unusual but valid user behaviors to handle
- **Out of scope**: Explicitly call out what this story does NOT cover
- **Dependencies**: Other stories, APIs, or team work this relies on
- **Open questions**: Decisions that need to be made before or during development

## Output Format

Structure as a numbered list of stories, each with: title, user story statement, acceptance criteria, story points, and notes. Format acceptance criteria as numbered Given/When/Then bullets. Keep the tone direct and technical — these are written for an engineering team.
