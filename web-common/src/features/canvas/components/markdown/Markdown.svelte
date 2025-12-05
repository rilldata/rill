<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import { queryServiceResolveTemplatedString } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createQuery } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { MarkdownCanvasComponent } from "./";
  import {
    getPositionClasses,
    hasTemplatingSyntax,
    formatResolvedContent,
    buildRequestBody,
  } from "./util";

  export let component: MarkdownCanvasComponent;

  $: specStore = component?.specStore;
  $: parent = component?.parent;
  $: markdownProperties = specStore ? $specStore : undefined;
  $: positionClasses = getPositionClasses(markdownProperties?.alignment);
  $: content = markdownProperties?.content ?? "";
  $: needsTemplating = hasTemplatingSyntax(content);
  $: applyFormatting = markdownProperties?.apply_formatting === true;

  $: timeAndFilterStore = component?.timeAndFilterStore;
  $: timeAndFilters = timeAndFilterStore ? $timeAndFilterStore : undefined;

  $: parentFilters = parent?.filters;
  $: parentWhereFilterStore = parentFilters?.whereFilter;
  $: parentDimensionThresholdStore = parentFilters?.dimensionThresholdFilters;
  $: globalWhereFilter =
    parentWhereFilterStore !== undefined ? $parentWhereFilterStore : undefined;
  $: globalDimensionThresholdFilters =
    parentDimensionThresholdStore !== undefined
      ? $parentDimensionThresholdStore
      : [];

  $: parentSpecStore = parent?.specStore;
  $: canvasData = parentSpecStore ? $parentSpecStore?.data : undefined;
  $: metricsViews = canvasData?.metricsViews ?? {};
  $: ({ instanceId } = $runtime);

  $: requestBody =
    needsTemplating && content && instanceId
      ? buildRequestBody({
          content,
          applyFormatting,
          timeRange: timeAndFilters?.timeRange,
          globalWhereFilter,
          globalDimensionThresholdFilters,
          metricsViews,
        })
      : null;

  $: resolveQuery = createQuery(
    {
      queryKey: [
        "resolveTemplatedString",
        instanceId,
        content,
        JSON.stringify(globalWhereFilter),
        JSON.stringify(globalDimensionThresholdFilters),
        JSON.stringify(timeAndFilters?.timeRange),
        Object.keys(metricsViews).join(","),
        applyFormatting,
      ],
      queryFn: async () => {
        if (!instanceId || !requestBody) return null;
        return await queryServiceResolveTemplatedString(
          instanceId,
          requestBody,
        );
      },
      enabled:
        needsTemplating &&
        !!requestBody &&
        !!instanceId &&
        !!requestBody.additionalTimeRange,
    },
    queryClient,
  );

  $: resolvedContent =
    needsTemplating && $resolveQuery?.data?.body
      ? applyFormatting && metricsViews && Object.keys(metricsViews).length > 0
        ? formatResolvedContent($resolveQuery.data.body, metricsViews)
        : $resolveQuery.data.body
      : content;

  $: renderPromise = marked(resolvedContent);
</script>

<div class="size-full px-2 overflow-y-auto select-text cursor-text bg-surface">
  <div class="canvas-markdown {positionClasses} h-full flex flex-col min-h-min">
    {#await renderPromise then htmlContent}
      {@html DOMPurify.sanitize(htmlContent)}
    {/await}
  </div>
</div>

<style lang="postcss">
  :global(.canvas-markdown) {
    @apply text-gray-900;
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
    @apply border-l-4 border-gray-300 pl-4 py-1 my-3 italic text-gray-600;
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
</style>
