<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import CustomChartRenderer from "@rilldata/web-common/features/components/charts/custom/CustomChartRenderer.svelte";
  import { substituteArgsRecursively } from "@rilldata/web-common/features/custom-viz/params";
  import { createQueryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import type { ComponentRefComponent } from "./index";

  export let component: ComponentRefComponent;
  export let editable: boolean = false;

  const client = useRuntimeClient();

  $: ({ specStore, timeAndFilterStore, resource } = component);

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

  $: metricsSQL = normalizeMetricsSQL(resolvedProps?.metrics_sql);
  $: vegaSpec = resolvedProps?.vega_spec as string | undefined;

  function normalizeMetricsSQL(value: unknown): string[] {
    if (typeof value === "string") return [value];
    if (Array.isArray(value)) {
      return value.filter((entry): entry is string => typeof entry === "string");
    }
    return [];
  }
</script>

{#if !componentSpec}
  <ComponentError error={m.canvas_component_ref_missing({ name: componentName })} />
{:else if renderer !== "custom_chart"}
  <ComponentError error={m.canvas_component_ref_unsupported_renderer({ renderer: renderer ?? "" })} />
{:else if $resolvedQuery.error && !optimisticProps}
  <ComponentError error={$resolvedQuery.error.message} />
{:else}
  <CustomChartRenderer
    name={component.id}
    spec={vegaSpec}
    whereFilter={$timeAndFilterStore?.where}
    timeRange={$timeAndFilterStore?.timeRange}
    {metricsSQL}
    showDataTable={editable}
  />
{/if}
