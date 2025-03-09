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
    @apply my-2;
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
  :global(.markdown a) {
    @apply text-blue-600;
  }
  :global(.markdown ul) {
    @apply list-disc pl-6 my-3;
  }
  :global(.markdown ol) {
    @apply list-decimal pl-6 my-3;
  }
  :global(.markdown li) {
    @apply text-sm my-1;
  }
  :global(.markdown blockquote) {
    @apply border-l-4 border-gray-300 pl-4 py-1 my-3 italic text-gray-600;
  }
  :global(.markdown code) {
    @apply bg-gray-100 px-1 py-0.5 rounded text-sm font-mono;
  }
  :global(.markdown pre) {
    @apply bg-gray-100 p-3 rounded my-3 overflow-x-auto;
  }
  :global(.markdown pre code) {
    @apply bg-transparent p-0 text-sm;
  }
  :global(.markdown hr) {
    @apply my-3 border-t border-gray-300 w-full;
  }
  :global(.markdown img) {
    @apply max-w-full h-auto my-3;
  }
</style>
