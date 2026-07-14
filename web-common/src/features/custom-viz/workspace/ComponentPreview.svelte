<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/custom/CustomChartRenderer.svelte";
  import {
    CONVERT_EXAMPLE_PROMPT,
    sendComponentFilePrompt,
  } from "@rilldata/web-common/features/custom-viz/component-ai-agent";
  import { substituteArgsRecursively } from "@rilldata/web-common/features/custom-viz/params";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { createQueryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { View, VisualizationSpec } from "svelte-vega";

  export let componentName: string;
  export let resource: V1Resource | undefined;
  export let args: Record<string, unknown>;
  // The component's file path; enables the "connect with AI" action in sample-data mode.
  export let filePath: string | undefined = undefined;

  const client = useRuntimeClient();

  const { developerChat } = featureFlags;

  let inlineViewVL: View;

  function convertWithAI() {
    if (!filePath) return;
    sendComponentFilePrompt(client, filePath, CONVERT_EXAMPLE_PROMPT);
  }

  function fixWithAI(error: string) {
    if (!filePath) return;
    sendComponentFilePrompt(
      client,
      filePath,
      `The component fails to render with this error: "${error}". Fix the file so it renders cleanly. Keep the structure valid: top-level custom_chart block, a trivial templated metrics_sql (SELECT params FROM {{ .params.metrics_view }} with ORDER BY and LIMIT; every field param referenced by the encodings or ORDER BY must be in the SELECT list; no CTEs/JOINs/subqueries/window functions/CASE/aggregates), typed role-named params, encoding types matching the bound param types (dimension → nominal/ordinal and never binned, time_dimension → temporal, measure → quantitative), no dataset-reshaping Vega transforms (only row-local calculate transforms the geometry needs), and no invented config properties.`,
    );
  }

  $: componentSpec =
    resource?.component?.state?.validSpec ?? resource?.component?.spec;
  $: renderer = componentSpec?.renderer;
  $: hasValidSpec = !!resource?.component?.state?.validSpec;

  // Server-side resolution (defaults merge, templating, vega param injection).
  // Only possible once the component has a valid spec; drafts fall back to
  // client-side substitution below.
  $: resolvedQuery = createQueryServiceResolveComponent(
    client,
    {
      component: componentName,
      args: args as PartialMessage<Struct>,
    },
    {
      query: {
        enabled: !!componentName && hasValidSpec,
        queryKey: [
          "resolve-component",
          client.instanceId,
          componentName,
          JSON.stringify(args),
          resource?.meta?.stateUpdatedOn ?? "",
        ],
      },
    },
  );

  $: optimisticProps = componentSpec?.rendererProperties
    ? (substituteArgsRecursively(componentSpec.rendererProperties, args) as
        | Record<string, unknown>
        | undefined)
    : undefined;

  $: resolvedProps = hasValidSpec
    ? (($resolvedQuery.data?.rendererProperties as
        | Record<string, unknown>
        | undefined) ?? optimisticProps)
    : optimisticProps;

  $: metricsSQL = normalizeMetricsSQL(resolvedProps?.metrics_sql);
  $: vegaSpec = resolvedProps?.vega_spec as string | undefined;

  // Sample-data mode: the spec carries inline data (e.g. imported from the Vega-Lite
  // examples gallery) and declares no queries, so it renders without a metrics view.
  $: inlineSpec = parseInlineDataSpec(vegaSpec, metricsSQL);

  function normalizeMetricsSQL(value: unknown): string[] {
    if (typeof value === "string") return [value];
    if (Array.isArray(value)) {
      return value.filter(
        (entry): entry is string =>
          typeof entry === "string" && entry.trim().length > 0,
      );
    }
    return [];
  }

  function parseInlineDataSpec(
    spec: string | undefined,
    metricsSQL: string[],
  ): VisualizationSpec | null {
    if (!spec || metricsSQL.length > 0) return null;
    try {
      const parsed = JSON.parse(spec);
      const hasInlineData =
        parsed?.data?.values !== undefined || parsed?.datasets !== undefined;
      return hasInlineData ? (parsed as VisualizationSpec) : null;
    } catch {
      return null;
    }
  }
</script>

<div class="size-full min-h-[400px] flex flex-col p-4">
  {#if !componentSpec}
    <ComponentError error={m.component_preview_no_spec()} />
  {:else if renderer !== "custom_chart"}
    <ComponentError
      error={m.canvas_component_ref_unsupported_renderer({
        renderer: renderer ?? "",
      })}
    />
  {:else if inlineSpec}
    <div
      class="mb-2 px-3 py-1.5 text-xs rounded-sm bg-surface-subtle text-fg-secondary w-fit flex items-center gap-x-3"
    >
      {m.component_preview_sample_data()}
      {#if $developerChat && filePath}
        <button
          class="text-primary-600 hover:text-primary-700 font-medium"
          onclick={convertWithAI}
        >
          {m.component_connect_with_ai()}
        </button>
      {/if}
    </div>
    <div class="flex-1 min-h-0 max-h-full overflow-hidden">
      <VegaLiteRenderer
        spec={inlineSpec}
        canvasDashboard
        config={getRillTheme(true)}
        bind:viewVL={inlineViewVL}
      />
    </div>
  {:else if $resolvedQuery.error && !optimisticProps}
    <div class="flex flex-col items-center justify-center gap-y-2 flex-1">
      <ComponentError error={$resolvedQuery.error.message} />
      {#if $developerChat && filePath}
        <button
          class="text-xs font-medium text-primary-600 hover:text-primary-700"
          onclick={() => fixWithAI($resolvedQuery.error?.message ?? "")}
        >
          {m.component_fix_with_ai()}
        </button>
      {/if}
    </div>
  {:else}
    <!-- overflow-hidden bounds charts with fixed/step sizes to the preview area. -->
    <div class="flex-1 min-h-[320px] max-h-full overflow-hidden">
      <CustomChartRenderer
        name={componentName}
        spec={vegaSpec}
        {metricsSQL}
        showDataTable
      />
    </div>
  {/if}
</div>
