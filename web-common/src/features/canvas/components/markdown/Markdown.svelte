<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import { createQuery } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { MarkdownCanvasComponent } from "./";
  import {
    getPositionClasses,
    hasTemplatingSyntax,
    formatResolvedContent,
    getResolveTemplatedStringQueryOptions,
  } from "./util";

  export let component: MarkdownCanvasComponent;

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

  const queryOptionsStore = getResolveTemplatedStringQueryOptions(component);
  $: resolveQuery = createQuery(queryOptionsStore, queryClient);

  // Store the last successfully resolved content to prevent flashing during refetches
  let lastResolvedContent: string | null = null;

  function applyFormattingIfNeeded(content: string): string {
    if (
      applyFormatting &&
      metricsViews &&
      Object.keys(metricsViews).length > 0
    ) {
      return formatResolvedContent(content, metricsViews);
    }
    return content;
  }

  $: resolvedContent = (() => {
    if (!needsTemplating) {
      lastResolvedContent = null;
      return content;
    }

    // If there's an error, return empty string so error message is shown instead
    if ($resolveQuery?.isError) {
      return "";
    }

    // If we're fetching and have a previous resolved value, use it to prevent flash
    if ($resolveQuery?.isFetching && lastResolvedContent) {
      return lastResolvedContent;
    }

    // Update stored resolved content when we have new data
    const queryData = $resolveQuery?.data;
    if (
      queryData &&
      typeof queryData === "object" &&
      queryData !== null &&
      "body" in queryData
    ) {
      const body = (queryData as { body?: string }).body;
      if (body) {
        const newResolvedContent = applyFormattingIfNeeded(body);
        lastResolvedContent = newResolvedContent;
        return newResolvedContent;
      }
    }

    return lastResolvedContent ?? content;
  })();
  $: renderPromise = marked(resolvedContent || "");

  $: errorMessage = (() => {
    const error = $resolveQuery?.error;
    if (!error) return "Failed to resolve template.";

    const err = error as any;
    return (
      err?.response?.data?.message ??
      err?.message ??
      "Failed to resolve template."
    );
  })();
</script>

<div class="size-full px-2 overflow-y-auto select-text cursor-text bg-surface">
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
  :global(.canvas-markdown) {
    @apply text-foreground;
  }
  :global(.canvas-markdown h1) {
    font-size: 24px;
    @apply font-medium;
  }
  :global(.canvas-markdown h2) {
    font-size: 20px;
    @apply font-medium;
  }
  :global(.canvas-markdown h3) {
    font-size: 18px;
    @apply font-medium;
  }
  :global(.canvas-markdown h4) {
    font-size: 16px;
    @apply font-medium;
  }
  :global(.canvas-markdown p) {
    font-size: 14px;
    @apply my-2;
  }

  :global(.canvas-markdown.items-center p) {
    @apply text-center w-full;
  }

  :global(.canvas-markdown.items-end p) {
    @apply text-right w-full;
  }
  :global(.canvas-markdown table) {
    @apply w-full border-collapse my-4;
  }
  :global(.canvas-markdown th) {
    @apply bg-gray-50 border px-4 py-2 text-left text-sm font-medium;
  }
  :global(.canvas-markdown td) {
    @apply border px-4 py-2 text-sm;
  }
  :global(.canvas-markdown tr:nth-child(even)) {
    @apply bg-gray-50;
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
    @apply text-sm my-1;
  }
  :global(.canvas-markdown blockquote) {
    @apply border-l-4 border-gray-300 pl-4 py-1 my-3 italic text-muted-foreground;
  }
  :global(.canvas-markdown code) {
    @apply bg-gray-100 px-1 py-0.5 rounded text-sm font-mono;
  }
  :global(.canvas-markdown pre) {
    @apply bg-gray-100 p-3 rounded my-3 overflow-x-auto;
  }
  :global(.canvas-markdown pre code) {
    @apply bg-transparent p-0 text-sm;
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
