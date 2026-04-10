# UX Copy Audit — Deferred Items

Items from the copy audit that need further discussion, design review, or broader codebase changes before implementation.

---

## Needs Design/Product Decision

### 1. "Widget" vs "Component" terminology in Canvas
- **Audit recommendation:** Standardize on one term
- **Why deferred:** This is a user-facing naming decision that touches many files (30+ in `web-common/src/features/canvas/`). Needs product alignment on which term resonates better with users. "Widget" is more end-user-friendly; "component" is used in code.
- **Files affected:** `CanvasBuilder.svelte`, all Canvas inspector files, Canvas component types

### 2. Canvas inspector copy: time filter overrides
- **Current:** "Overriding inherited time filters from canvas." / "Overrides inherited time filters from canvas when ON."
- **Suggested:** "Using this component's own time range instead of the canvas default." / "When on, this component uses its own time range."
- **Why deferred:** These are in the Canvas inspector which has specific UX patterns; need to verify the rewrites fit the visual context.

### 3. Chart sort label: "delta" vs "change"
- **Current:** "Y-axis delta ascending"
- **Suggested:** "Y-axis change (ascending)"
- **Why deferred:** "Delta" may be intentional domain terminology for analytics users. Needs user research input.

### 4. Onboarding: "Model sources" step label
- **Current:** "Import data" → "Model sources" → "Define metrics" → "Explore insights"
- **Suggested:** "Import data" → "Transform with SQL" → "Define metrics" → "Explore insights"
- **Why deferred:** This changes the conceptual framing of a core workflow step. Needs product alignment.

### 5. Canvas creation dialog copy
- **Current:** "Which metrics view should this dashboard reference?" / "This will determine the measures and dimensions you can explore on this dashboard."
- **Suggested:** "Choose a metrics view for this dashboard" / "This determines which measures and dimensions are available."
- **Why deferred:** Dialog copy changes should be verified in context with the visual layout.

---

## Needs Broader Codebase Sweep

### 6. Standardize "Failed to" / "Unable to" → "Couldn't" across all error toasts
- **Status:** Applied to 9 specific inline error messages. ~100+ more instances exist across `web-common` and `web-admin` in toast notifications and API error handlers.
- **Why deferred:** Many of these are in notification/toast utilities that may have structured patterns. A bulk find-and-replace risks breaking error handler logic or test assertions. Needs a dedicated pass with test verification.
- **Files to audit:** `web-admin/src/features/projects/`, `web-admin/src/features/organizations/`, notification helpers

### 7. Notification toast cleanup
- **Current issues identified:**
  - "Remote project changes fetched and merged." → too technical
  - "Remote project changes fetched and merged. Your changes have been stashed." → Git jargon
  - "Triggered an ad-hoc run of this report." → formal
  - `Queried ${table} in workspace` → developer-facing
  - "Converted filter type to Select" → exposes implementation
- **Why deferred:** These are in notification/toast utilities that may have test coverage. Need to verify each change doesn't break anything.

### 8. Billing copy: "Input a valid" pattern
- **Current:** "Input a valid payment to maintain access." / "Input a valid billing address to maintain access."
- **Suggested:** "Add a valid payment method to keep access." / "Add a billing address to keep access."
- **Why deferred:** Billing copy changes need careful review — these may be legally reviewed strings.

### 9. "Hibernating" terminology in trial expiration
- **Current:** "Your trial has expired and this org's projects are now hibernating."
- **Suggested:** "Your trial has expired. Projects are paused and dashboards are unavailable until you upgrade."
- **Why deferred:** "Hibernating" is a Rill-specific state term used in the backend. Changing the user-facing label needs alignment with how the feature is documented elsewhere.

---

## Strategic / Long-term

### 10. Voice & tone guidelines
- The codebase has at least 3 distinct voices (casual onboarding, neutral product UI, formal errors). Recommend extending the onboarding warmth into the product UI.
- **Action:** Define and document a Rill voice & tone guide as part of the design system.

### 11. Copy centralization
- Strings are scattered across ~150+ files with no centralization.
- **Action:** Consider per-feature constants files (some features like `dimension-filters/constants.ts` already do this well). Not full i18n, just copy co-location.

### 12. `name-utils.ts` validation message
- **Current:** `Filename cannot contain special characters like /, <, >, :, ", \, |, ?, or *. Please choose a different name.`
- **Suggested:** "Names can only contain letters, numbers, hyphens, and underscores."
- **Why deferred:** The suggested rewrite changes the validation framing from "disallowed chars" to "allowed chars". Need to verify this accurately describes the actual validation logic (it may allow more chars than letters/numbers/hyphens/underscores).

### 13. `error-message-helpers.ts` field validation
- **Current:** `` Selected ${fieldLabel}: "${field}" is not valid. ``
- **Suggested:** `` "${field}" isn't a valid ${fieldLabel}. It may have been renamed or removed. ``
- **Why deferred:** This is a generic helper used across multiple surfaces. The "renamed or removed" guidance may not apply in all contexts. Needs per-usage audit.

### 14. Explore dashboard compare column copy
- **Current:** "No comparison dimension selected"
- **Suggested:** "No comparison selected"
- **Why deferred:** Minor change, but it's in a dense data visualization context. Dropping "dimension" might reduce clarity for power users.
