<script lang="ts">
  import type { MarkdownCanvasComponent } from ".";
  import Markdown from "./Markdown.svelte";
  import MarkdownProviderBackend from "./MarkdownProviderBackend.svelte";
  import { hasGoTemplateExpressions } from "./util";

  export let component: MarkdownCanvasComponent;

  $: ({
    specStore,
    timeAndFilterStore,
    parent: { name: canvasName },
  } = component);
  $: spec = $specStore;
  $: hasExpressions = hasGoTemplateExpressions(spec.content);
  $: markdownProperties = {
    content: spec.content,
    alignment: spec.alignment,
  };
</script>

{#if hasExpressions}
  <MarkdownProviderBackend {spec} {timeAndFilterStore} {canvasName} />
{:else}
  <Markdown {markdownProperties} />
{/if}
