---
name: google-stitch
description: Use Google Stitch MCP tools for UI design work — generating mockups, prototyping screens, extracting design tokens, creating design systems, and producing responsive variants. Trigger when the user asks to design, mockup, prototype, or explore UI ideas, or when working on any visual/frontend design task.
---

# Google Stitch — AI UI Design Tool

Google Stitch is a Gemini-powered UI design tool from Google Labs. Use it for rapid visual exploration, mockups, design systems, and prototyping. It is available via MCP tools in this workspace.

## When to Use Stitch

- User asks to **design**, **mockup**, or **prototype** a screen or page
- User wants to **explore design ideas** or **generate variations**
- User needs **design tokens** extracted from an existing design
- User wants **responsive variants** (mobile/desktop/tablet) of a screen
- User is doing **design system** work (creating, applying, or auditing consistency)
- User wants to **generate design assets** (logos, icons, hero images)
- User is working on **multi-screen flows** (onboarding, wizards, etc.)

## Tool Reference

### Screen Generation

| Tool | Use When |
|---|---|
| `mcp__stitch__generate_screen_from_text` | Generating a single screen from a text description |
| `mcp__stitch__batch_generate_screens` | Generating multiple related screens in one operation |
| `mcp__stitch__edit_screens` | Modifying an existing generated screen |
| `mcp__stitch__generate_variants` | Creating multiple design variations of a screen |
| `mcp__stitch__orchestrate_design` | Full orchestration: assets + UI in one prompt |

### Design System & Tokens

| Tool | Use When |
|---|---|
| `mcp__stitch__create_design_system` | Setting up colors, typography, roundness, and theme |
| `mcp__stitch__update_design_system` | Modifying an existing design system |
| `mcp__stitch__apply_design_system` | Applying a design system to screen generation |
| `mcp__stitch__generate_design_tokens` | Exporting tokens as CSS variables, Tailwind, SCSS, or JSON |
| `mcp__stitch__extract_design_context` | Extracting design DNA (colors, typography, spacing) from a screen |
| `mcp__stitch__generate_style_guide` | Creating a visual style guide from an existing screen |

### Responsive & Trends

| Tool | Use When |
|---|---|
| `mcp__stitch__generate_responsive_variant` | Creating mobile/desktop/tablet versions of a screen |
| `mcp__stitch__suggest_trending_design` | Applying design trends (glassmorphism, bento-grid, etc.) |

### Assets

| Tool | Use When |
|---|---|
| `mcp__stitch__generate_design_asset` | Generating logos, icons, illustrations, hero images |
| `mcp__stitch__extract_components` | Extracting reusable components from a screen |

### Code & Export

| Tool | Use When |
|---|---|
| `mcp__stitch__fetch_screen_code` | Getting the HTML/CSS/React code for a screen |
| `mcp__stitch__fetch_screen_image` | Getting a screenshot/image of a screen |

### Project Management

| Tool | Use When |
|---|---|
| `mcp__stitch__create_project` | Starting a new design project |
| `mcp__stitch__list_projects` | Listing existing projects |
| `mcp__stitch__get_project` | Getting details of a specific project |
| `mcp__stitch__list_screens` | Listing all screens in a project |
| `mcp__stitch__get_screen` | Getting details of a specific screen |
| `mcp__stitch__get_workspace_project` | Checking if a project is linked to the current workspace |
| `mcp__stitch__set_workspace_project` | Linking a project to the current workspace |

### Quality & Analysis

| Tool | Use When |
|---|---|
| `mcp__stitch__analyze_accessibility` | Checking accessibility of a generated design |
| `mcp__stitch__compare_designs` | Comparing two screens side by side |

## Workflow

### Quick Mockup
1. `create_project` (or use existing)
2. `generate_screen_from_text` with a descriptive prompt
3. `fetch_screen_image` to preview
4. `fetch_screen_code` to get the code reference

### Design System First
1. `create_project`
2. `create_design_system` with Rill's brand colors, fonts, and roundness
3. `update_design_system` to apply it
4. `apply_design_context` when generating screens to maintain consistency

### Multi-Screen Flow
1. `create_project`
2. `batch_generate_screens` with all screens described
3. Review with `fetch_screen_image` for each
4. `extract_design_context` from the best screen, then `apply_design_context` to refine others

### Export for Development
1. `generate_design_tokens` → get CSS variables or Tailwind config
2. `fetch_screen_code` → get HTML/CSS/React reference code
3. Translate React output to Svelte for the Rill codebase

## Important Notes

- **Model tiers**: Gemini 3.1 Pro (higher quality, 50/month) or Gemini 3 Flash (faster, 350/month)
- **Device types**: MOBILE, DESKTOP, TABLET
- **Code output is React/HTML/CSS** — for Rill's Svelte codebase, use Stitch for visual reference and design exploration, then translate to Svelte
- **Design tokens** can be exported in formats directly usable in Rill: CSS variables or Tailwind config
- Screens can take a few minutes to generate — do not retry on timeout; check with `get_screen` later
