<script lang="ts">
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { createRuntimeServiceRenderMarkdown } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get, type Readable } from "svelte/store";
  import type { MarkdownSpec } from ".";
  import { Markdown } from ".";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";

  export let spec: MarkdownSpec;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  export let canvasName: string;

  $: ({ instanceId } = $runtime);
  $: ctx = getCanvasStore(canvasName, instanceId);
  $: ({
    metricsView: { getMeasureForMetricView },
  } = ctx.canvasEntity);

  // Create mutation once
  const renderMarkdown = createRuntimeServiceRenderMarkdown();

  // Trigger mutation when dependencies change
  let lastKey = "";
  $: {
    const {
      timeRange: { timeZone, start, end },
      where,
    } = $timeAndFilterStore;
    const timeRange = start && end ? { start, end } : undefined;
    const { content } = spec;

    const key = JSON.stringify({
      instanceId,
      content,
      where,
      timeRange,
      timeZone,
    });

    if (instanceId && content && key !== lastKey) {
      lastKey = key;

      $renderMarkdown.mutate({
        instanceId,
        data: {
          template: content,
          where,
          timeRange,
          timeZone,
        },
      });
    }
  }

  function formatRenderedMarkdown(markdown: string): string {
    const tokenRegex = /__RILL_FORMAT__([^:]+)::([^:]+)::(.+?)__END__/g;

    return markdown.replace(
      tokenRegex,
      (match, metricsView, measureOrDim, rawValue) => {
        try {
          const measureStore = getMeasureForMetricView(
            measureOrDim,
            metricsView,
          );
          const measureSpec = get(measureStore);

          if (measureSpec) {
            const formatter = createMeasureValueFormatter(measureSpec);
            const numValue = parseFloat(rawValue);
            if (!isNaN(numValue)) {
              return formatter(numValue);
            }
          }
        } catch (e) {
          // Ignore formatting errors
        }

        return rawValue;
      },
    );
  }

  $: renderedContent = (() => {
    const result = $renderMarkdown;

    if (result.isError) {
      return `**Error rendering markdown:**\n\n${result.error?.message || "Unknown error"}`;
    }

    if (result.isPending || !result.isSuccess) {
      return "Loading...";
    }

    const rawMarkdown = result.data?.renderedMarkdown || spec.content;
    return formatRenderedMarkdown(rawMarkdown);
  })();

  $: markdownProperties = {
    content: renderedContent,
    alignment: spec.alignment,
  };
</script>

<Markdown {markdownProperties} />
