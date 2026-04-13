# AI-Assisted Error Resolution — Phase 1a Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an "Explain this error" CTA to every error displayed in Rill Developer (code view and visual editors) that opens the AI chat panel with a pre-composed prompt containing error context.

**Architecture:** A shared `composeErrorPrompt()` utility builds a structured prompt from error metadata (message, file path, line number, surrounding code). An `ExplainErrorButton` Svelte component wraps this utility and calls `sidebarActions.startChat(prompt)` to open the AI chat panel and auto-send the message. This button is added to `WorkspaceEditorContainer` (code view errors), `ModelWorkspace` (model table errors), `ExploreWorkspace` (visual error page), `CanvasWorkspace` (canvas loading errors), and `SubmissionError` (connector form errors).

**Tech Stack:** Svelte 4/5, TypeScript, lucide-svelte (SparklesIcon), existing `sidebarActions.startChat()` API

---

## File Structure

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `web-common/src/features/chat/error-prompt-composer.ts` | Builds structured prompts from error context |
| Create | `web-common/src/features/chat/error-prompt-composer.spec.ts` | Unit tests for prompt composer |
| Create | `web-common/src/features/chat/ExplainErrorButton.svelte` | Reusable "Explain this error" CTA button |
| Modify | `web-common/src/layout/workspace/WorkspaceEditorContainer.svelte` | Add ExplainErrorButton to the error banner |
| Modify | `web-common/src/features/workspaces/ModelWorkspace.svelte` | Add ExplainErrorButton to model table error area |
| Modify | `web-common/src/features/workspaces/ExploreWorkspace.svelte` | Add ExplainErrorButton to explore visual error page |
| Modify | `web-common/src/features/workspaces/CanvasWorkspace.svelte` | Pass error context to CanvasLoadingState for CTA |
| Modify | `web-common/src/features/canvas/CanvasLoadingState.svelte` | Add ExplainErrorButton when errorMessage is shown |
| Modify | `web-common/src/components/forms/SubmissionError.svelte` | Add ExplainErrorButton to connector form errors |

---

### Task 1: Error Prompt Composer Utility

**Files:**
- Create: `web-common/src/features/chat/error-prompt-composer.ts`
- Create: `web-common/src/features/chat/error-prompt-composer.spec.ts`

- [ ] **Step 1: Write the failing tests**

Create `web-common/src/features/chat/error-prompt-composer.spec.ts`:

```typescript
import { describe, it, expect } from "vitest";
import { composeErrorPrompt } from "./error-prompt-composer";

describe("composeErrorPrompt", () => {
  it("includes error message, file path, and resource type", () => {
    const result = composeErrorPrompt({
      errorMessage: "unexpected token at line 5",
      filePath: "/models/my_model.sql",
      fileContent: "SELECT *\nFROM table\nWHERE x = 1\nAND y = 2\nORDER BY z\nLIMIT 10",
    });

    expect(result).toContain("unexpected token at line 5");
    expect(result).toContain("/models/my_model.sql");
    expect(result).toContain("SQL model");
    expect(result).toContain("Please explain what's wrong and suggest how to fix it.");
  });

  it("includes line number and surrounding context when available", () => {
    const lines = Array.from({ length: 20 }, (_, i) => `line ${i + 1}`);
    const result = composeErrorPrompt({
      errorMessage: "syntax error",
      filePath: "/models/test.sql",
      fileContent: lines.join("\n"),
      lineNumber: 10,
    });

    expect(result).toContain("Line 10");
    // Should include 5 lines above (5-9) and 5 below (11-15)
    expect(result).toContain("line 5");
    expect(result).toContain("line 15");
    // Should NOT include distant lines
    expect(result).not.toContain("line 1\n");
    expect(result).not.toContain("line 20");
  });

  it("includes whole file if short and no line number", () => {
    const content = "SELECT *\nFROM table\nWHERE x = 1";
    const result = composeErrorPrompt({
      errorMessage: "error",
      filePath: "/models/short.sql",
      fileContent: content,
    });

    expect(result).toContain("SELECT *");
    expect(result).toContain("WHERE x = 1");
  });

  it("includes first 30 + last 10 lines for long files without line number", () => {
    const lines = Array.from({ length: 80 }, (_, i) => `line ${i + 1}`);
    const result = composeErrorPrompt({
      errorMessage: "error",
      filePath: "/models/long.sql",
      fileContent: lines.join("\n"),
    });

    expect(result).toContain("line 1");
    expect(result).toContain("line 30");
    expect(result).toContain("line 71");
    expect(result).toContain("line 80");
    expect(result).not.toContain("line 40");
  });

  it("detects resource types from file path", () => {
    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/metrics/mv.yaml",
        fileContent: "version: 1",
      }),
    ).toContain("metrics view");

    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/dashboards/canvas.yaml",
        fileContent: "type: canvas",
      }),
    ).toContain("canvas dashboard");

    expect(
      composeErrorPrompt({
        errorMessage: "e",
        filePath: "/connectors/pg.yaml",
        fileContent: "driver: postgres",
      }),
    ).toContain("connector");
  });

  it("appends additional error count", () => {
    const result = composeErrorPrompt({
      errorMessage: "first error",
      filePath: "/models/test.sql",
      fileContent: "SELECT 1",
      additionalErrorCount: 3,
    });

    expect(result).toContain("There are also 3 other errors in this file.");
  });

  it("strips credential-like fields from connector file content", () => {
    const content = [
      "driver: postgres",
      "host: db.example.com",
      "password: secret123",
      "token: abc-def-ghi",
      "secret: mysecret",
      "api_key: key123",
      "port: 5432",
    ].join("\n");

    const result = composeErrorPrompt({
      errorMessage: "connection failed",
      filePath: "/connectors/pg.yaml",
      fileContent: content,
    });

    expect(result).not.toContain("secret123");
    expect(result).not.toContain("abc-def-ghi");
    expect(result).not.toContain("mysecret");
    expect(result).not.toContain("key123");
    expect(result).toContain("host: db.example.com");
    expect(result).toContain("port: 5432");
  });

  it("keeps prompt under 2000 characters for typical errors", () => {
    const result = composeErrorPrompt({
      errorMessage: "unexpected token near 'SELECT'",
      filePath: "/models/orders.sql",
      fileContent: Array.from({ length: 30 }, (_, i) => `-- line ${i + 1}: some SQL code here`).join("\n"),
      lineNumber: 15,
    });

    expect(result.length).toBeLessThan(2000);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd web-common && npx vitest run src/features/chat/error-prompt-composer.spec.ts`
Expected: FAIL — module `./error-prompt-composer` not found

- [ ] **Step 3: Implement the prompt composer**

Create `web-common/src/features/chat/error-prompt-composer.ts`:

```typescript
/**
 * Composes a structured prompt from error context for the AI developer agent.
 */

const CREDENTIAL_PATTERNS =
  /^(\s*)(password|secret|token|api_key|apikey|api_secret|access_key|private_key|client_secret|auth_token|credentials|connection_string)(\s*:\s*)(.+)$/gim;

const MAX_PROMPT_LENGTH = 2000;
const SHORT_FILE_THRESHOLD = 50;
const CONTEXT_LINES_ABOVE = 5;
const CONTEXT_LINES_BELOW = 5;
const LONG_FILE_HEAD = 30;
const LONG_FILE_TAIL = 10;

export interface ErrorPromptInput {
  errorMessage: string;
  filePath: string;
  fileContent?: string | null;
  lineNumber?: number;
  additionalErrorCount?: number;
}

export function composeErrorPrompt(input: ErrorPromptInput): string {
  const {
    errorMessage,
    filePath,
    fileContent,
    lineNumber,
    additionalErrorCount,
  } = input;

  const resourceType = inferResourceType(filePath);
  const isConnector = resourceType === "connector";

  const parts: string[] = [];

  // Header
  parts.push(
    `I have an error in my ${resourceType} file \`${filePath}\`${lineNumber ? ` at Line ${lineNumber}` : ""}:`,
  );
  parts.push("");

  // Error message
  parts.push(`**Error:** ${errorMessage}`);
  parts.push("");

  // File context
  if (fileContent) {
    const sanitized = isConnector
      ? stripCredentials(fileContent)
      : fileContent;
    const context = extractContext(sanitized, lineNumber);
    if (context) {
      parts.push("**Relevant code:**");
      parts.push("```");
      parts.push(context);
      parts.push("```");
      parts.push("");
    }
  }

  // Additional errors
  if (additionalErrorCount && additionalErrorCount > 0) {
    parts.push(
      `There are also ${additionalErrorCount} other errors in this file.`,
    );
    parts.push("");
  }

  // Closing
  parts.push("Please explain what's wrong and suggest how to fix it.");

  let prompt = parts.join("\n");

  // Truncate if over limit; trim the code context
  if (prompt.length > MAX_PROMPT_LENGTH && fileContent) {
    const withoutCode = parts
      .filter((p) => !p.startsWith("```") && !p.startsWith("**Relevant"))
      .join("\n");
    if (withoutCode.length < MAX_PROMPT_LENGTH) {
      prompt = withoutCode;
    }
  }

  return prompt;
}

function inferResourceType(filePath: string): string {
  if (filePath.endsWith(".sql")) return "SQL model";
  if (filePath.startsWith("/models/")) return "SQL model";
  if (filePath.startsWith("/sources/")) return "source";
  if (filePath.startsWith("/connectors/")) return "connector";

  // For YAML, use directory to determine type
  if (filePath.startsWith("/metrics/")) return "metrics view";
  if (filePath.startsWith("/dashboards/")) return "canvas dashboard";
  if (filePath.startsWith("/explores/")) return "explore dashboard";
  if (filePath.startsWith("/apis/")) return "API";
  if (filePath.startsWith("/themes/")) return "theme";

  if (filePath.endsWith(".yaml") || filePath.endsWith(".yml")) return "YAML";
  return "file";
}

function stripCredentials(content: string): string {
  return content.replace(
    CREDENTIAL_PATTERNS,
    "$1$2$3[REDACTED]",
  );
}

function extractContext(
  content: string,
  lineNumber: number | undefined,
): string {
  const lines = content.split("\n");

  if (lineNumber) {
    const idx = lineNumber - 1; // 0-based
    const start = Math.max(0, idx - CONTEXT_LINES_ABOVE);
    const end = Math.min(lines.length, idx + CONTEXT_LINES_BELOW + 1);
    return lines
      .slice(start, end)
      .map((line, i) => {
        const num = start + i + 1;
        const marker = num === lineNumber ? " >" : "  ";
        return `${marker} ${num}: ${line}`;
      })
      .join("\n");
  }

  if (lines.length <= SHORT_FILE_THRESHOLD) {
    return lines.map((line, i) => `  ${i + 1}: ${line}`).join("\n");
  }

  // Long file: first N + last M
  const head = lines
    .slice(0, LONG_FILE_HEAD)
    .map((line, i) => `  ${i + 1}: ${line}`);
  const tail = lines
    .slice(-LONG_FILE_TAIL)
    .map((line, i) => `  ${lines.length - LONG_FILE_TAIL + i + 1}: ${line}`);
  return [...head, "  ...", ...tail].join("\n");
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd web-common && npx vitest run src/features/chat/error-prompt-composer.spec.ts`
Expected: All 8 tests PASS

- [ ] **Step 5: Commit**

```bash
git add web-common/src/features/chat/error-prompt-composer.ts web-common/src/features/chat/error-prompt-composer.spec.ts
git commit -m "feat: add error prompt composer for AI-assisted error resolution"
```

---

### Task 2: ExplainErrorButton Component

**Files:**
- Create: `web-common/src/features/chat/ExplainErrorButton.svelte`

- [ ] **Step 1: Create the ExplainErrorButton component**

Create `web-common/src/features/chat/ExplainErrorButton.svelte`:

```svelte
<script lang="ts">
  import { SparklesIcon } from "lucide-svelte";
  import { sidebarActions } from "./layouts/sidebar/sidebar-store";
  import {
    composeErrorPrompt,
    type ErrorPromptInput,
  } from "./error-prompt-composer";

  export let errorMessage: string;
  export let filePath: string;
  export let fileContent: string | null | undefined = undefined;
  export let lineNumber: number | undefined = undefined;
  export let additionalErrorCount: number | undefined = undefined;

  function handleClick() {
    const prompt = composeErrorPrompt({
      errorMessage,
      filePath,
      fileContent,
      lineNumber,
      additionalErrorCount,
    });
    sidebarActions.startChat(prompt);
  }
</script>

<button
  class="explain-error-btn"
  on:click|stopPropagation={handleClick}
  aria-label="Explain this error with AI"
  title="Explain this error"
>
  <SparklesIcon size="12px" />
  <span>Explain this error</span>
</button>

<style lang="postcss">
  .explain-error-btn {
    @apply inline-flex items-center gap-1 px-2 py-0.5;
    @apply text-[11px] font-medium;
    @apply text-accent-primary-action hover:text-fg-accent;
    @apply bg-transparent hover:bg-surface-hover;
    @apply rounded-sm cursor-pointer;
    @apply border border-transparent hover:border-accent-primary-action/30;
    @apply transition-colors duration-150;
    @apply flex-shrink-0;
  }

  .explain-error-btn:focus-visible {
    @apply outline-none ring-1 ring-accent-primary-action ring-offset-1;
  }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/chat/ExplainErrorButton.svelte
git commit -m "feat: add ExplainErrorButton component"
```

---

### Task 3: Add "Explain this error" to WorkspaceEditorContainer (Code View)

This is the primary error banner shown beneath the code editor for all resource types (models, metrics views, canvases, explores, generic files). It's rendered by `WorkspaceEditorContainer.svelte`.

**Files:**
- Modify: `web-common/src/layout/workspace/WorkspaceEditorContainer.svelte`

- [ ] **Step 1: Read the current file**

Read `web-common/src/layout/workspace/WorkspaceEditorContainer.svelte` in full to confirm current state matches expectations. It currently shows the error in a red banner at the bottom with a `CancelCircle` icon and the error text.

- [ ] **Step 2: Add ExplainErrorButton to the error banner**

The existing error banner (lines 49-58) renders `effectiveError` text. Add an `ExplainErrorButton` after the error text, and pass the file path and error context. The component already has `parseError` (which includes `startLocation.line`) and `resource` props. We need to also accept `filePath` and `fileContent` to pass through.

Edit `web-common/src/layout/workspace/WorkspaceEditorContainer.svelte`:

1. Add props for `filePath` and `fileContent`:

After line 19 (`export let remoteContent`), add:
```svelte
  export let filePath: string | undefined = undefined;
  export let fileContent: string | null | undefined = undefined;
```

2. Add import for ExplainErrorButton (after other imports, around line 3):
```svelte
  import ExplainErrorButton from "@rilldata/web-common/features/chat/ExplainErrorButton.svelte";
```

3. Derive the line number from the parse error:
After line 33 (`$: derivedError = parseError?.message ?? rootCauseReconcileError;`), add:
```svelte
  $: errorLineNumber = parseError?.startLocation?.line;
```

4. Replace the error banner's inner `<div>` (lines 55-57) to add the button:
```svelte
      <div class="flex gap-x-2 items-center justify-between">
        <div class="flex gap-x-2 items-center min-w-0">
          <CancelCircle className="text-destructive flex-shrink-0" /><span class="break-words">{effectiveError}</span>
        </div>
        {#if filePath}
          <ExplainErrorButton
            errorMessage={effectiveError ?? ""}
            {filePath}
            {fileContent}
            lineNumber={errorLineNumber}
          />
        {/if}
      </div>
```

- [ ] **Step 3: Update callers to pass filePath and fileContent**

Now we need each workspace that uses `WorkspaceEditorContainer` to pass `filePath` and `fileContent`. Check each caller:

**MetricsWorkspace.svelte** (line 111-115): Already has `filePath` in scope and `$remoteContent`. Add:
```svelte
<WorkspaceEditorContainer
  {resource}
  {parseError}
  remoteContent={$remoteContent}
  {filePath}
  fileContent={$remoteContent}
>
```

**CanvasWorkspace.svelte** (line 122-126): Already has `filePath` and `$remoteContent`. Add:
```svelte
<WorkspaceEditorContainer
  resource={data}
  {parseError}
  remoteContent={$remoteContent}
  {filePath}
  fileContent={$remoteContent}
>
```

**ExploreWorkspace.svelte** (line 123-127): Has `filePath` and `$remoteContent`. Add:
```svelte
<WorkspaceEditorContainer
  resource={exploreResource ?? metricsViewResource}
  {parseError}
  remoteContent={$remoteContent}
  {filePath}
  fileContent={$remoteContent}
>
```

**ModelWorkspace.svelte** (line 136): Uses `WorkspaceEditorContainer` without any error props — the model workspace renders errors separately in the table area (lines 152-164). The editor container here has NO error/resource/parseError props. We still want to support it if errors are added later, but don't force it now. Skip this one.

**+page.svelte (generic file view)** (line 102-107): Has `path` and `$remoteContent`. Add:
```svelte
<WorkspaceEditorContainer
  slot="body"
  {resource}
  {parseError}
  remoteContent={$remoteContent}
  filePath={path}
  fileContent={$remoteContent}
>
```

- [ ] **Step 4: Commit**

```bash
git add web-common/src/layout/workspace/WorkspaceEditorContainer.svelte \
  web-common/src/features/workspaces/MetricsWorkspace.svelte \
  web-common/src/features/workspaces/CanvasWorkspace.svelte \
  web-common/src/features/workspaces/ExploreWorkspace.svelte \
  web-local/src/routes/\(application\)/\(workspace\)/files/\[...file\]/+page.svelte
git commit -m "feat: add 'Explain this error' CTA to code view error banner"
```

---

### Task 4: Add "Explain this error" to Model Workspace Table Errors

The model workspace renders errors in its own table area (not through `WorkspaceEditorContainer`). These are shown in the `<svelte:fragment slot="error">` block.

**Files:**
- Modify: `web-common/src/features/workspaces/ModelWorkspace.svelte`

- [ ] **Step 1: Read the current file and add ExplainErrorButton**

The error block is at lines 151-165. Add the button after the error list.

1. Add import (with the other imports near the top):
```svelte
import ExplainErrorButton from "@rilldata/web-common/features/chat/ExplainErrorButton.svelte";
```

2. Get file content from the fileArtifact. Add after line 38 (`remoteContent`):
Already destructured `remoteContent` at line 38.

3. Modify the error block. Replace lines 151-165 with:
```svelte
        <svelte:fragment slot="error">
          {#if allErrors.length > 0}
            <div
              transition:slide={{ duration: 200 }}
              class="border border-destructive bg-destructive/15 dark:bg-destructive/30 text-fg-primary border-l-4 px-2 py-5 max-h-72 overflow-auto flex flex-col gap-2"
              aria-label="Model errors"
            >
              {#each allErrors as error (error.message)}
                <div>
                  {getUserFriendlyError(error.message ?? "")}
                </div>
              {/each}
              <div class="flex justify-end pt-1">
                <ExplainErrorButton
                  errorMessage={getUserFriendlyError(allErrors[0]?.message ?? "")}
                  {filePath}
                  fileContent={$remoteContent}
                  lineNumber={allErrors[0]?.startLocation?.line}
                  additionalErrorCount={allErrors.length > 1 ? allErrors.length - 1 : undefined}
                />
              </div>
            </div>
          {/if}
        </svelte:fragment>
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/workspaces/ModelWorkspace.svelte
git commit -m "feat: add 'Explain this error' CTA to model table error area"
```

---

### Task 5: Add "Explain this error" to Explore Visual Error Page

When the explore workspace is in "viz" mode and there's an error, it renders an `ErrorPage` component (line 136-139 in ExploreWorkspace.svelte). Add the button there.

**Files:**
- Modify: `web-common/src/features/workspaces/ExploreWorkspace.svelte`

- [ ] **Step 1: Add the button below the ErrorPage**

1. Add import:
```svelte
import ExplainErrorButton from "@rilldata/web-common/features/chat/ExplainErrorButton.svelte";
```

2. Replace lines 136-140 with:
```svelte
              {#if parseError || rootCauseReconcileError}
                <div class="flex flex-col items-center gap-4">
                  <ErrorPage
                    body={parseError?.message ?? rootCauseReconcileError ?? ""}
                    fatal
                    header="Unable to load dashboard preview"
                    statusCode={404}
                  />
                  <ExplainErrorButton
                    errorMessage={parseError?.message ?? rootCauseReconcileError ?? ""}
                    {filePath}
                    fileContent={$remoteContent}
                    lineNumber={parseError?.startLocation?.line}
                  />
                </div>
```

- [ ] **Step 2: Commit**

```bash
git add web-common/src/features/workspaces/ExploreWorkspace.svelte
git commit -m "feat: add 'Explain this error' CTA to explore visual error page"
```

---

### Task 6: Add "Explain this error" to Canvas Loading State Errors

When the canvas workspace is in "viz" mode and there's an error, it renders via `CanvasLoadingState` with `errorMessage`. We need to add the button there.

**Files:**
- Modify: `web-common/src/features/canvas/CanvasLoadingState.svelte`
- Modify: `web-common/src/features/workspaces/CanvasWorkspace.svelte`

- [ ] **Step 1: Read CanvasLoadingState.svelte**

Read `web-common/src/features/canvas/CanvasLoadingState.svelte` to understand its structure and where the error message is displayed.

- [ ] **Step 2: Add ExplainErrorButton to CanvasLoadingState**

Add new props `filePath` and `fileContent` to `CanvasLoadingState.svelte`, and render the button when there's an error message.

1. Add import:
```svelte
import ExplainErrorButton from "@rilldata/web-common/features/chat/ExplainErrorButton.svelte";
```

2. Add props:
```svelte
export let filePath: string | undefined = undefined;
export let fileContent: string | null | undefined = undefined;
```

3. Where the `errorMessage` is displayed, add the button adjacent to it.

- [ ] **Step 3: Pass filePath and fileContent from CanvasWorkspace**

In `CanvasWorkspace.svelte`, update the `CanvasLoadingState` usage (around line 135-140):
```svelte
<CanvasLoadingState
  {ready}
  {isReconciling}
  {isLoading}
  errorMessage={rootCauseReconcileError}
  {filePath}
  fileContent={$remoteContent}
>
```

- [ ] **Step 4: Commit**

```bash
git add web-common/src/features/canvas/CanvasLoadingState.svelte \
  web-common/src/features/workspaces/CanvasWorkspace.svelte
git commit -m "feat: add 'Explain this error' CTA to canvas loading error state"
```

---

### Task 7: Add "Explain this error" to Connector Form Errors

The `SubmissionError` component is used in connector/source forms. It shows connection test failures. We add an optional "Explain this error" button.

**Files:**
- Modify: `web-common/src/components/forms/SubmissionError.svelte`

- [ ] **Step 1: Add optional ExplainErrorButton to SubmissionError**

The connector form is a modal, not a file workspace, so the file path context is different. We'll make the button optional; it shows only when `filePath` is provided.

1. Add import and props:
```svelte
import ExplainErrorButton from "@rilldata/web-common/features/chat/ExplainErrorButton.svelte";

export let filePath: string | undefined = undefined;
```

2. Add the button after the message div (inside the `flex-1 min-w-0` div, after the existing content). Before the closing `</div>` of the `flex-1` wrapper, add:
```svelte
      {#if filePath}
        <div class="mt-2">
          <ExplainErrorButton
            errorMessage={message}
            {filePath}
          />
        </div>
      {/if}
```

Note: We intentionally do NOT pass `fileContent` for connector forms in this implementation — the connector YAML may contain credentials, and the prompt composer will only have the error message to work with. The developer agent can read the file itself (it has file read tools) and will not expose credentials in the chat.

- [ ] **Step 2: Pass filePath from AddDataForm or callers where available**

Look at where `SubmissionError` is used in the connector flow and pass `filePath` if the connector resource has one. This is optional; if the connector file doesn't exist yet (first-time setup), the button simply won't appear.

Search for `SubmissionError` usage and update callers that have a file path in scope.

- [ ] **Step 3: Commit**

```bash
git add web-common/src/components/forms/SubmissionError.svelte
git commit -m "feat: add optional 'Explain this error' CTA to connector form errors"
```

---

### Task 8: Verify and Polish

- [ ] **Step 1: Run the unit tests**

Run: `cd web-common && npx vitest run src/features/chat/error-prompt-composer.spec.ts`
Expected: All tests pass

- [ ] **Step 2: Run the full frontend test suite to check for regressions**

Run: `npm run test -w web-common`
Expected: No regressions in existing tests

- [ ] **Step 3: Run lint/format**

Run: `npm run quality`
Expected: No lint errors

- [ ] **Step 4: Final commit with any fixes**

If quality check required changes:
```bash
git add -u
git commit -m "style: fix lint/format issues"
```

- [ ] **Step 5: Commit summary**

The implementation should have these commits:
1. `feat: add error prompt composer for AI-assisted error resolution`
2. `feat: add ExplainErrorButton component`
3. `feat: add 'Explain this error' CTA to code view error banner`
4. `feat: add 'Explain this error' CTA to model table error area`
5. `feat: add 'Explain this error' CTA to explore visual error page`
6. `feat: add 'Explain this error' CTA to canvas loading error state`
7. `feat: add optional 'Explain this error' CTA to connector form errors`
