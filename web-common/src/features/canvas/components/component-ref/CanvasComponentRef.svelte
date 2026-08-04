<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import FlintChartRenderer from "@rilldata/web-common/features/components/charts/flint/FlintChartRenderer.svelte";
  import type { FlintChartSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
  import { substituteArgsRecursively } from "@rilldata/web-common/features/custom-viz/params";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { createQueryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import type { ComponentRefComponent } from "./index";

  export let component: ComponentRefComponent;
  export let editable: boolean = false;

  const client = useRuntimeClient();

  $: ({ specStore, timeAndFilterStore, resource } = component);

  $: ({
    parent: { theme },
  } = component);

  // Theme mode (light/dark) is separate from which theme is selected.
  let themeMode: "light" | "dark";
  $: themeMode = $themeControl === "dark" ? "dark" : "light";
  $: currentTheme = $theme?.resolvedThemeObject?.[themeMode];

  $: componentName = $specStore.component;
  $: args = component.args($specStore);
  $: componentSpec =
    $resource?.component?.state?.validSpec ??
    (component.parent.allowUnvalidatedSpec
      ? $resource?.component?.spec
      : undefined);
  $: renderer = componentSpec?.renderer;

  // Server-side resolution merges declared defaults, resolves {{ .params.* }}
  // templating (plus .env/.user) and injects scalar params as native Vega-Lite params.
  $: resolvedQuery = createQueryServiceResolveComponent(
    client,
    {
      component: componentName,
      args: args as PartialMessage<Struct>,
    },
    {
      query: {
        enabled: !!componentName && !!componentSpec,
        queryKey: [
          "resolve-component",
          client.instanceId,
          componentName,
          JSON.stringify(args),
          $resource?.meta?.stateUpdatedOn ?? "",
        ],
      },
    },
  );

  // While the server resolution is in flight (or failed transiently), fall back to
  // optimistic client-side substitution of the raw renderer properties.
  $: optimisticProps = componentSpec?.rendererProperties
    ? (substituteArgsRecursively(componentSpec.rendererProperties, args) as
        | Record<string, unknown>
        | undefined)
    : undefined;

  $: resolvedProps =
    ($resolvedQuery.data?.rendererProperties as
      | Record<string, unknown>
      | undefined) ?? optimisticProps;

  // Flint takes a single row set, so components declare one metrics_sql query.
  $: metricsSQL = normalizeMetricsSQL(resolvedProps?.metrics_sql);
  $: flintSpec = resolvedProps?.spec as FlintChartSpec | undefined;

  $: metricsViewName = $specStore.metrics_view as string | undefined;
  $: timeGrain = $timeAndFilterStore?.timeGrain;

  function normalizeMetricsSQL(value: unknown): string | undefined {
    if (typeof value === "string") return value;
    if (Array.isArray(value)) {
      // Multi-query components are a Vega-era shape that Flint cannot express.
      return value.find((entry): entry is string => typeof entry === "string");
    }
    return undefined;
  }
</script>

{#if !componentSpec}
  <ComponentError
    error={m.canvas_component_ref_missing({ name: componentName })}
  />
{:else if renderer !== "custom_chart"}
  <ComponentError
    error={m.canvas_component_ref_unsupported_renderer({
      renderer: renderer ?? "",
    })}
  />
{:else if $resolvedQuery.error && !optimisticProps}
  <ComponentError error={$resolvedQuery.error.message} />
{:else}
  <FlintChartRenderer
    name={component.id}
    spec={flintSpec}
    {metricsSQL}
    {metricsViewName}
    {timeGrain}
    whereFilter={$timeAndFilterStore?.where}
    timeRange={$timeAndFilterStore?.timeRange}
    showDataTable={editable}
    {themeMode}
    theme={currentTheme}
  />
{/if}
