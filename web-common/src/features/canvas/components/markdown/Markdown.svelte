<script lang="ts">
  import type { MarkdownProperties } from "@rilldata/web-common/features/templates/types";
  import type { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import DOMPurify from "dompurify";
  import { marked } from "marked";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  $: markdownProperties = rendererProperties as MarkdownProperties;
  $: css = markdownProperties.css || {};

  $: styleString = Object.entries(css)
    .map(([k, v]) => `${k}:${v}`)
    .join(";");
</script>

<div
  class="markdown size-full p-2 flex flex-col justify-center bg-white"
  style={styleString}
>
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
