<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import { createQuery } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { MarkdownCanvasComponent } from "./";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import {
    getPositionClasses,
    hasTemplatingSyntax,
    formatResolvedContent,
    unformatResolvedContent,
    getResolveTemplatedStringQueryOptions,
  } from "./util";

  export let component: MarkdownCanvasComponent;

  const runtimeClient = useRuntimeClient();

  $: specStore = component?.specStore;
  $: spec = specStore ? $specStore : undefined;
  $: content = spec?.content ?? "";
  $: applyFormatting = spec?.apply_formatting === true;
  $: positionClasses = getPositionClasses(spec?.alignment);
  $: needsTemplating = hasTemplatingSyntax(content);

  $: parentSpecStore = component?.parent?.specStore;
  $: metricsViews = parentSpecStore
    ? ($parentSpecStore?.data?.metricsViews ?? {})
    : {};

  const queryOptionsStore = getResolveTemplatedStringQueryOptions(
    component,
    runtimeClient,
  );
  $: resolveQuery = createQuery(queryOptionsStore, queryClient);

  // Store the last successfully resolved body to prevent flashing during refetches
  let lastResolvedBody: string | null = null;

  // The resolved markdown before measure formatting:
  // with formatting on, the runtime wraps each measure value in a format token.
  $: resolvedBody = (() => {
    if (!needsTemplating) {
      lastResolvedBody = null;
      return content;
    }

    // If there's an error, return empty string so error message is shown instead
    if ($resolveQuery?.isError) {
      return "";
    }

    // If we're fetching and have a previous resolved value, use it to prevent flash
    if ($resolveQuery?.isFetching && lastResolvedBody) {
      return lastResolvedBody;
    }

    // Update stored resolved body when we have new data
    const queryData = $resolveQuery?.data;
    if (
      queryData &&
      typeof queryData === "object" &&
      queryData !== null &&
      "body" in queryData
    ) {
      const body = (queryData as { body?: string }).body;
      // An empty body is a valid result: the template rendered nothing.
      if (typeof body === "string") {
        lastResolvedBody = body;
        return body;
      }
    }

    // Until the template resolves there is nothing to show:
    // the raw template source is not meant to be read.
    return lastResolvedBody ?? "";
  })();
  $: resolvedContent =
    applyFormatting && Object.keys(metricsViews).length > 0
      ? formatResolvedContent(resolvedBody, metricsViews)
      : resolvedBody;
  $: renderPromise = marked(resolvedContent || "");

  $: errorMessage = (() => {
    const error = $resolveQuery?.error;
    if (!error) return "Failed to resolve template.";

    return extractErrorMessage(error);
  })();

  // The resolved text for the workbook and CSV exports; see BaseCanvasComponent.exportPayload.
  $: component?.exportPayload.set(
    resolvedContent
      ? {
          kind: "text",
          text: resolvedContent,
          unformattedText: unformatResolvedContent(resolvedBody),
        }
      : undefined,
  );
</script>

<div
  class="size-full px-2 overflow-y-auto select-text cursor-text bg-surface-card"
>
  <div class="canvas-markdown {positionClasses} h-full flex flex-col min-h-min">
    {#if needsTemplating && $resolveQuery?.isError}
      <div class="markdown-error">
        <p>{errorMessage}</p>
      </div>
    {:else}
      {#await renderPromise then htmlContent}
        {@html DOMPurify.sanitize(htmlContent)}
      {/await}
    {/if}
  </div>
</div>

<style lang="postcss">
  /* Every font size is multiplied by `--text-scale`, the component's text
     size preset, set on its card. Text without a rule of its own, such as
     h5 and h6, takes the scaled base size. */
  :global(.canvas-markdown) {
    font-size: calc(12px * var(--text-scale, 1));
    @apply text-fg-primary;
  }
  :global(.canvas-markdown h1) {
    font-size: calc(24px * var(--text-scale, 1));
    @apply font-medium;
  }
  :global(.canvas-markdown h2) {
    font-size: calc(20px * var(--text-scale, 1));
    @apply font-medium;
  }
  :global(.canvas-markdown h3) {
    font-size: calc(18px * var(--text-scale, 1));
    @apply font-medium;
  }
  :global(.canvas-markdown h4) {
    font-size: calc(16px * var(--text-scale, 1));
    @apply font-medium;
  }
  :global(.canvas-markdown p) {
    font-size: calc(14px * var(--text-scale, 1));
    @apply my-2;
  }

  :global(.canvas-markdown.items-center p) {
    @apply text-center w-full;
  }

  :global(.canvas-markdown.items-end p) {
    @apply text-right w-full;
  }
  /* Tailwind's `text-sm`, scaled. */
  :global(.canvas-markdown th),
  :global(.canvas-markdown td),
  :global(.canvas-markdown li),
  :global(.canvas-markdown code) {
    font-size: calc(0.875rem * var(--text-scale, 1));
    line-height: calc(1.25rem * var(--text-scale, 1));
  }
  :global(.canvas-markdown table) {
    @apply w-full border-collapse my-4;
  }
  :global(.canvas-markdown th) {
    @apply bg-surface-background border px-4 py-2 text-left font-medium;
  }
  :global(.canvas-markdown td) {
    @apply border px-4 py-2;
  }
  :global(.canvas-markdown tr:nth-child(even)) {
    @apply bg-surface-background;
  }
  :global(.canvas-markdown tr:hover) {
    @apply bg-gray-100;
  }
  :global(.canvas-markdown a) {
    @apply text-blue-600;
  }
  :global(.canvas-markdown ul) {
    @apply list-disc pl-6 my-3;
  }
  :global(.canvas-markdown ol) {
    @apply list-decimal pl-6 my-3;
  }
  :global(.canvas-markdown li) {
    @apply my-1;
  }
  :global(.canvas-markdown blockquote) {
    @apply border-l-4 border-gray-300 pl-4 py-1 my-3 italic text-fg-secondary;
  }
  :global(.canvas-markdown code) {
    @apply bg-gray-100 px-1 py-0.5 rounded font-mono;
  }
  :global(.canvas-markdown pre) {
    @apply bg-gray-100 p-3 rounded my-3 overflow-x-auto;
  }
  :global(.canvas-markdown pre code) {
    @apply bg-transparent p-0;
  }
  :global(.canvas-markdown hr) {
    @apply my-3 border-t border-gray-300 w-full;
  }
  :global(.canvas-markdown img) {
    @apply max-w-full h-auto my-3;
  }
  .markdown-error {
    @apply bg-red-50 border border-red-200 rounded px-3 py-2 my-2 text-sm text-red-700;
  }
</style>
