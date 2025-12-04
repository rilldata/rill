<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import type { MarkdownCanvasComponent } from "./";
  import { getPositionClasses } from "./util";

  export let component: MarkdownCanvasComponent;

  $: ({ specStore } = component);
  $: markdownProperties = $specStore;

  $: positionClasses = getPositionClasses(markdownProperties.alignment);
</script>

<div class="size-full px-2 overflow-y-auto select-text cursor-text bg-surface">
  <div class="canvas-markdown {positionClasses} h-full flex flex-col min-h-min">
    {#await marked(markdownProperties.content) then content}
      {@html DOMPurify.sanitize(content)}
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
