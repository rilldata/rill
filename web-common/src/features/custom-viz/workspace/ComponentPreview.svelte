<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import FlintChartRenderer from "@rilldata/web-common/features/components/charts/flint/FlintChartRenderer.svelte";
  import type { FlintChartSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
  import { sendComponentFilePrompt } from "@rilldata/web-common/features/custom-viz/component-ai-agent";
  import { optimisticRendererProps } from "@rilldata/web-common/features/custom-viz/params";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    V1TimeGrain,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { createQueryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let componentName: string;
  export let resource: V1Resource | undefined;
  export let args: Record<string, unknown>;
  // The component's file path; enables the "fix with AI" action.
  export let filePath: string | undefined = undefined;

  const client = useRuntimeClient();

  const { developerChat } = featureFlags;

  function fixWithAI(error: string) {
    if (!filePath) return;
    sendComponentFilePrompt(
      client,
      filePath,
      `It fails to render with this error: "${error}". Fix it so it renders cleanly, keeping the chart it draws, its bound channels, and whichever spec format it already uses.`,
    );
  }

  // The preview has no canvas parent, so there is no theme object to apply; only
  // the light/dark mode, which the chart's colors resolve from CSS variables.
  let themeMode: "light" | "dark";
  $: themeMode = $themeControl === "dark" ? "dark" : "light";

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

  $: optimisticProps = optimisticRendererProps(
    componentSpec?.rendererProperties,
    args,
  );

  $: resolvedProps = hasValidSpec
    ? (($resolvedQuery.data?.rendererProperties as
        | Record<string, unknown>
        | undefined) ?? optimisticProps)
    : optimisticProps;

  $: metricsSQL = normalizeMetricsSQL(resolvedProps?.metrics_sql);
  $: flintSpec = resolvedProps?.spec as FlintChartSpec | undefined;
  // An ejected component carries the Vega-Lite it used to compile to instead of a chart spec.
  $: vegaSpec = resolvedProps?.vega_spec as string | undefined;

  // The metrics view the preview is bound to, from whichever param declares it.
  $: metricsViewParam = componentSpec?.params?.find(
    (param) => param.type === "metrics_view",
  );
  $: metricsViewName = (
    metricsViewParam?.name
      ? (args[metricsViewParam.name] ?? metricsViewParam.default)
      : undefined
  ) as string | undefined;

  function normalizeMetricsSQL(value: unknown): string | undefined {
    if (typeof value === "string") return value.trim() ? value : undefined;
    if (Array.isArray(value)) {
      return value.find(
        (entry): entry is string =>
          typeof entry === "string" && entry.trim().length > 0,
      );
    }
    return undefined;
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
  {:else if !resolvedProps}
    <!-- The properties reference something only the server resolves, so either its response or a
         reconcile of the edited file is still outstanding; both settle on their own. -->
    <div class="flex-1 flex items-center justify-center">
      <ReconcilingSpinner />
    </div>
  {:else}
    <!-- overflow-hidden bounds charts with fixed/step sizes to the preview area. -->
    <div class="flex-1 min-h-[320px] max-h-full overflow-hidden">
      <!-- The preview has no dashboard time controls to inherit a grain from, so it buckets by day:
           enough to keep a time series readable, where the raw timestamps would render one mark per event. -->
      <FlintChartRenderer
        name={componentName}
        spec={flintSpec}
        {vegaSpec}
        {metricsSQL}
        {metricsViewName}
        timeGrain={V1TimeGrain.TIME_GRAIN_DAY}
        showDataTable
        {themeMode}
      />
    </div>
  {/if}
</div>
