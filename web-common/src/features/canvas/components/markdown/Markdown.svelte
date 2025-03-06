<script lang="ts">
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import type { MarkdownSpec } from "./";
  import { getPositionClasses } from "./util";

  export let rendererProperties: V1ComponentSpecRendererProperties;
  $: markdownProperties = rendererProperties as MarkdownSpec;

  $: positionClasses = getPositionClasses(markdownProperties.alignment);
</script>

<div class="size-full p-2 bg-white overflow-y-auto">
  <div class="markdown {positionClasses} h-full flex flex-col min-h-min">
    {#await marked(markdownProperties.content) then content}
      {@html DOMPurify.sanitize(content)}
    {/await}
  </div>
</div>

<style lang="postcss">
  :global(.markdown) {
    @apply text-gray-800;
  }
  :global(.markdown h1) {
    font-size: 24px;
    @apply font-medium;
  }
  :global(.markdown h2) {
    font-size: 20px;
    @apply font-medium;
  }
  :global(.markdown h3) {
    font-size: 18px;
    @apply font-medium;
  }
  :global(.markdown h4) {
    font-size: 16px;
    @apply font-medium;
  }
  :global(.markdown p) {
    font-size: 14px;
  }
  :global(.markdown table) {
    @apply w-full border-collapse my-4;
  }
  :global(.markdown th) {
    @apply bg-gray-50 border border-gray-200 px-4 py-2 text-left text-sm font-medium;
  }
  :global(.markdown td) {
    @apply border border-gray-200 px-4 py-2 text-sm;
  }
  :global(.markdown tr:nth-child(even)) {
    @apply bg-gray-50;
  }
  :global(.markdown tr:hover) {
    @apply bg-gray-100;
  }
</style>
