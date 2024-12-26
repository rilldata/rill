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

  $: content = markdownProperties.content || "";
</script>

<div
  class="markdown size-full items-center flex justify-center bg-white"
  style={styleString}
>
  {#await marked(content) then parsedContent}
    {@html DOMPurify.sanitize(parsedContent)}
  {/await}
</div>

<style lang="postcss">
  :global(.markdown h1) {
    font-size: 2em;
    font-weight: 500;
  }
</style>
