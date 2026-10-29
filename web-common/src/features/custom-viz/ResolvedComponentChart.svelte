<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import FlintChartRenderer from "@rilldata/web-common/features/components/charts/flint/FlintChartRenderer.svelte";
  import type { FlintChartSpec } from "@rilldata/web-common/features/custom-viz/flint/compile";
  import {
    normalizeMetricsSQL,
    optimisticRendererProps,
  } from "@rilldata/web-common/features/custom-viz/params";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1ComponentSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQueryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";

  // Renders a standalone component bound to the given args: the server resolves its templated
  // renderer properties, and until it responds the args are substituted client-side.
  // Props not declared here are passed through to FlintChartRenderer.

  export let componentName: string;
  export let componentSpec: V1ComponentSpec | undefined;
  // Whether the server can resolve the component, i.e. it has a valid spec.
  export let resolvable: boolean;
  export let args: Record<string, unknown>;
  // Invalidates the resolved properties when the component file changes.
  export let stateUpdatedOn: string | undefined;
  // Shown when there is no spec to render.
  export let missingError: string;
  export let metricsViewName: string | undefined;

  const client = useRuntimeClient();

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
        enabled: !!componentName && resolvable,
        queryKey: [
          "resolve-component",
          client.instanceId,
          componentName,
          JSON.stringify(args),
          stateUpdatedOn ?? "",
        ],
      },
    },
  );

  // While the server resolution is in flight (or failed transiently), fall back to
  // optimistic client-side substitution of the raw renderer properties.
  $: optimisticProps = optimisticRendererProps(
    componentSpec?.rendererProperties,
    args,
  );

  $: resolvedProps = resolvable
    ? (($resolvedQuery.data?.rendererProperties as
        | Record<string, unknown>
        | undefined) ?? optimisticProps)
    : optimisticProps;

  // Flint takes a single row set, so components declare one metrics_sql query.
  $: metricsSQL = normalizeMetricsSQL(resolvedProps?.metrics_sql);
  $: flintSpec = resolvedProps?.spec as FlintChartSpec | undefined;
  // An ejected component carries the Vega-Lite it used to compile to instead of a chart spec.
  $: vegaSpec = resolvedProps?.vega_spec as string | undefined;
</script>

{#if !componentSpec}
  <ComponentError error={missingError} />
{:else if renderer !== "custom_chart"}
  <ComponentError
    error={m.canvas_component_ref_unsupported_renderer({
      renderer: renderer ?? "",
    })}
  />
{:else if $resolvedQuery.error && !optimisticProps}
  <slot name="error" error={$resolvedQuery.error.message}>
    <ComponentError error={$resolvedQuery.error.message} />
  </slot>
{:else if !resolvedProps}
  <!-- The properties reference something only the server resolves, so either its response or a
       reconcile of the edited file is still outstanding; both settle on their own. -->
  <div class="size-full flex-1 flex items-center justify-center">
    <ReconcilingSpinner />
  </div>
{:else}
  <FlintChartRenderer
    spec={flintSpec}
    {vegaSpec}
    {metricsSQL}
    {metricsViewName}
    {...$$restProps}
  />
{/if}
