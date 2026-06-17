---
description: Create a complete Go-to-Market (GTM) plan for a product, feature, or launch — including positioning, channels, messaging, timeline, and success metrics
allowed-tools: Read, AskUserQuestion, WebSearch
argument-hint: "<product, feature, or launch to create a GTM plan for>"
---

Build a complete Go-to-Market plan covering positioning, audience targeting, channel strategy, messaging, launch timeline, and success metrics.

Input: $ARGUMENTS

## Instructions

### 1. Gather Launch Context

If not provided, use `AskUserQuestion` to ask:
- What are we launching? (new product, feature, pricing change, expansion)
- Who is the primary target audience?
- What is the launch date or target window?
- What are the top 1–2 business goals for this launch?
- Any known constraints (budget, team, regions, existing commitments)?

### 2. Positioning Statement

Write a positioning statement using the Geoffrey Moore format:

> For [target customer] who [has this need or problem], [product/feature name] is a [category] that [key benefit]. Unlike [primary alternative], our solution [key differentiator].

Test the positioning against these questions:
- Is the "unlike" contrast meaningful to the customer (not just internally)?
- Does the key benefit tie directly to a measurable outcome?
- Is the category framing right — too narrow (limits growth) or too broad (loses clarity)?

### 3. Audience Segmentation

Define:
- **Primary audience**: Who this launch is designed for first. Be specific.
- **Secondary audience**: Who else benefits, and how messaging differs for them.
- **Influencers and champions**: Who drives adoption within a company or community?

### 4. Messaging Framework

Create a messaging hierarchy:
- **One-liner** (10 words max): Used in headlines, subject lines, social
- **Elevator pitch** (2–3 sentences): Used in sales, demos, outbound
- **Full value proposition** (1 paragraph): Used in landing pages, press materials

For each audience segment, note any messaging adjustments.

### 5. Channel Strategy

For each channel, specify: goal, tactic, owner, and expected impact.

Evaluate which channels to use based on where the audience lives:
- **Owned**: Blog, in-app notifications, email newsletter, changelog
- **Earned**: Press, analyst briefings, community word-of-mouth, influencer reviews
- **Paid**: Ads, sponsored content, events
- **Partner**: Co-marketing, integrations announcements, referrals

### 6. Launch Timeline

Create a phased timeline:
- **T-4 weeks**: Internal alignment, sales enablement, beta/preview customers
- **T-2 weeks**: Soft launch to waitlist or early adopters, collect feedback
- **T-0 (Launch day)**: Public announcement, coordinated across all channels
- **T+2 weeks**: Follow-up content, case studies, retargeting

### 7. Success Metrics

Define:
- **Primary metric**: The one number that signals launch success
- **Leading indicators**: What to watch in the first 48–72 hours
- **Lagging indicators**: What to evaluate at 30/60/90 days
- **Failure signal**: At what threshold do we escalate or pivot?

## Output Format

Structure the output as a standalone GTM brief in clean markdown — concise enough to share with leadership, detailed enough for the team to execute from. Use tables for the channel strategy and timeline.
