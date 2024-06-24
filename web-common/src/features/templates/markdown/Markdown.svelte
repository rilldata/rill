<script lang="ts">
  import DOMPurify from "dompurify";
  import { marked } from "marked";
  import { MarkdownProperties } from "@rilldata/web-common/features/templates/types";
  import { V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";

  export let rendererProperties: V1ComponentSpecRendererProperties;

  $: kpiProperties = rendererProperties as MarkdownProperties;
  $: css = kpiProperties.css || {};

  $: styleString = Object.entries(css)
    .map(([k, v]) => `${k}:${v}`)
    .join(";");
</script>

<div
  class="markdown size-full items-center flex justify-center"
  style={styleString}
>
  {#await marked(kpiProperties.content) then content}
    {@html DOMPurify.sanitize(content)}
  {/await}
</div>
