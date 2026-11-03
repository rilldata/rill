# Looker Competitive Analysis (Aug 2026)

Research basis: cloud.google.com/looker docs + pricing + release notes, Next '25/'26 coverage, G2, third-party pricing teardowns, and the Rill codebase + docs.rilldata.com. Structure mirrors the other `audit/` competitor analyses.

**TL;DR:** Looker in 2026 is a three-time Gartner Leader accelerating hard on AI (Conversational Analytics GA, CA API + A2A, managed MCP server, Dashboard Agents/Agentic Workflows previews, Code Interpreter GA) while pruning everything else (Modeler shelved, Studio-in-Looker killed, Looker Studio renamed back to Data Studio, customer-hosted abandoned) — and bleeding customers to a migration wave that Omni and Lightdash monetize today. Its unfixable weaknesses are Rill's strengths: warehouse-bound dashboard latency (~38% of G2 reviewers complain), LookML burden, $60K–200K/yr platform fees, Gemini-only AI with overage billing starting Oct 1, 2026. Rill's plan: get on the migration shortlist (`rill migrate lookml` + content), match the shortlisted Gemini features (chat API, verified answers, agentic monitoring), and close joins/SQL-API/enterprise gaps — selling "everything you wanted Looker to be: sub-second, local-first, agent-native, a tenth of the price."

| File | Contents |
|---|---|
| [`looker-features-and-market.md`](looker-features-and-market.md) | LookML semantic layer (joins/symmetric aggregates, PDTs, aggregate awareness, CI, VS Code ext), BI product + Data Studio split, Gemini AI timeline (CA GA, CA API, MCP, agents), embedded, pricing (incl. third-party platform-fee figures), Gartner standing, migration-wave evidence, mid-2025→mid-2026 ship log |
| [`rill-feature-inventory.md`](rill-feature-inventory.md) | Rill current state, Looker-angled: speed-vs-PDT story, DX, AI/MCP, embedded economics — plus gaps (joins, delivery destinations, ecosystem, SCIM/audit) and migration-play assets |
| [`comparison.md`](comparison.md) | Feature matrix across AI, semantic layer, BI, embedded, enterprise/pricing + structural advantages/weaknesses both ways |
| [`roadmap.md`](roadmap.md) | Phase 1 migration play (`rill migrate lookml`, destinations), Phase 2 Gemini-neutralizers (chat API, verified queries, evals, agentic monitoring, BYO-LLM timing vs Oct 2026 overage billing), Phase 3 structural (joins, SQL API, GCP Marketplace, enterprise pack), GTM, non-goals |
