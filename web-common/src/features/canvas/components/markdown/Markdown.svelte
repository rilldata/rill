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

<div class="{positionClasses} markdown size-full p-2 flex flex-col bg-white">
  {#await marked(markdownProperties.content) then content}
    {@html DOMPurify.sanitize(content)}
  {/await}
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
</style>
