<script lang="ts">
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { Editor } from "@tiptap/core";
  import { Button } from "@rilldata/web-common/components/button";
  import { getAdvancedFilterEditorPlugins } from "@rilldata/web-common/features/dashboards/filters/advanced-filters/plugins.svelte.ts";
  import { onMount } from "svelte";
  import { convertHTMLToSql } from "@rilldata/web-common/features/dashboards/filters/advanced-filters/advanced-filter.ts";
  import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";

  let {
    expressionFilterManager,
  }: {
    expressionFilterManager: ExpressionFilterManager;
  } = $props();

  let firstMetricsView = $derived(
    expressionFilterManager.metricsViewsProvider.metricsViewNames[0] ?? "",
  );

  let element: HTMLDivElement;
  let editor: Editor;

  let draftHtml = $state("");

  function onSubmit() {
    const sql = convertHTMLToSql(draftHtml.trim());
    try {
      const { expr } = convertFilterParamToExpression(sql);
      console.log(sql, expr);
      expressionFilterManager.setExprForMetricsView(firstMetricsView, expr);
    } catch (e) {
      console.log(e);
      // TODO: show error
    }
  }

  onMount(() => {
    editor = new Editor({
      element,
      extensions: getAdvancedFilterEditorPlugins({
        expressionFilterManager,
        dimensions: expressionFilterManager.metricsViewsProvider.dimensions,
        measures: expressionFilterManager.metricsViewsProvider.measures,
      }),
      content: "",
      onTransaction: () => {
        // force re-render so `editor.isActive` works as expected
        editor = editor;
      },
      onUpdate: ({ editor }) => {
        draftHtml = editor.getText();
      },
    });

    return () => {
      editor.destroy();
    };
  });
</script>

<div class="chat-input-form">
  <div class="chat-input-container" bind:this={element}></div>
  <Button onClick={onSubmit}>Submit</Button>
</div>

<style lang="postcss">
  .chat-input-form {
    @apply flex flex-row items-center pb-2;
  }

  :global(.tiptap) {
    @apply outline-none;
    @apply text-sm leading-relaxed;
  }

  .chat-input-container {
    @apply w-full max-h-32 overflow-auto;
  }

  :global(.tiptap p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    @apply text-fg-muted pointer-events-none absolute;
  }
</style>
