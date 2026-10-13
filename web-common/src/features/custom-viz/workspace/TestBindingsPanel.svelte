<script lang="ts">
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useDefaultMetrics } from "@rilldata/web-common/features/canvas/selector";
  import {
    dimensionRole,
    getDeclaredParams,
    isOrderByParam,
    orderByOptions,
    paramChannels,
    reconcileOrderByArg,
  } from "@rilldata/web-common/features/custom-viz/params";
  import { pickTestArgs } from "@rilldata/web-common/features/custom-viz/test-args";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    type V1ComponentParam,
    type V1MetricsViewSpec,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { Writable } from "svelte/store";
  import ParamValueInput from "./ParamValueInput.svelte";

  export let resource: V1Resource | undefined;
  export let argsStore: Writable<Record<string, unknown>>;

  const client = useRuntimeClient();

  $: params = getDeclaredParams(resource, true);

  $: metricsViewsQuery = useFilteredResources(client, ResourceKind.MetricsView);
  $: metricsViewResources = ($metricsViewsQuery?.data ?? []).filter(
    (res) => !!res.metricsView?.state?.validSpec,
  );
  $: metricsViewNames = metricsViewResources
    .map((res) => res.meta?.name?.name)
    .filter((name): name is string => !!name);

  $: defaultMetricsQuery = useDefaultMetrics(client);
  $: defaultMetrics = $defaultMetricsQuery?.data;

  $: rendererProperties = (
    resource?.component?.state?.validSpec ?? resource?.component?.spec
  )?.rendererProperties as Record<string, unknown> | undefined;

  $: channels = paramChannels(rendererProperties);
  $: chartType = (
    rendererProperties?.spec as { chartType?: string } | undefined
  )?.chartType;

  // How many distinct values suit each dimension param, shown so a rebinding keeps the chart
  // readable. Derived reactively: a template function call would not track channels or chartType.
  $: hints = Object.fromEntries(
    params.map((param) => [
      param.name ?? "",
      dimensionHint(param, channels, chartType),
    ]),
  );

  function dimensionHint(
    param: V1ComponentParam,
    channels: Record<string, string>,
    chartType: string | undefined,
  ): string | undefined {
    if (param.type !== "dimension") return undefined;
    switch (dimensionRole(param.name, channels, chartType)) {
      case "series":
        return m.component_hint_series();
      case "pair":
        return m.component_hint_pair();
      case "detail":
        return m.component_hint_detail();
      default:
        return undefined;
    }
  }

  // Seed test bindings for params without a value, picking fields that suit the channel each param
  // is drawn on. The pick reads the metrics view's data, so a newer seed supersedes a pending one.
  let seedRun = 0;
  $: void seedDefaults(params, rendererProperties, defaultMetrics);

  async function seedDefaults(
    params: V1ComponentParam[],
    rendererProperties: Record<string, unknown> | undefined,
    defaultMetrics:
      | { metricsViewName: string; metricsViewSpec?: V1MetricsViewSpec }
      | null
      | undefined,
  ) {
    if (!params.length || !defaultMetrics?.metricsViewSpec) return;
    const run = ++seedRun;
    const defaults = await pickTestArgs(client, params, rendererProperties, {
      name: defaultMetrics.metricsViewName,
      spec: defaultMetrics.metricsViewSpec,
    });
    if (run !== seedRun) return;
    argsStore.update((args) => {
      const next = { ...args };
      for (const [key, value] of Object.entries(defaults)) {
        if (next[key] === undefined) next[key] = value;
      }
      return next;
    });
  }

  // The metrics view spec bound to a field param (through its metrics_view param).
  function boundMetricsViewSpec(
    param: V1ComponentParam,
    args: Record<string, unknown>,
  ) {
    const mvParam = param.metricsViewParam;
    const name = mvParam ? args[mvParam] : undefined;
    if (typeof name !== "string") return undefined;
    return metricsViewResources.find((res) => res.meta?.name?.name === name)
      ?.metricsView?.state?.validSpec;
  }

  function setArg(name: string | undefined, value: unknown) {
    if (!name) return;
    argsStore.update((args) => {
      const next = { ...args };
      if (value === undefined) {
        delete next[name];
      } else {
        next[name] = value;
      }
      // Changing a metrics view invalidates field bindings that depend on it.
      for (const param of params) {
        if (param.metricsViewParam === name && param.name) {
          delete next[param.name];
        }
      }
      // order_by follows the field it pointed at, or re-seeds from the measure.
      reconcileOrderByArg(params, next, {
        name,
        previousValue: args[name],
      });
      return next;
    });
  }
</script>

<div class="flex flex-col">
  {#if params.length === 0}
    <div class="px-5 py-4 text-xs text-fg-secondary">
      {m.component_no_params()}
    </div>
  {:else}
    <div
      class="px-5 pt-3 pb-1 text-[11px] font-semibold uppercase text-fg-secondary"
    >
      {m.component_test_values()}
    </div>
    <div class="px-5 pb-2 text-[11px] text-fg-secondary">
      {m.component_test_values_note()}
    </div>
    {#each params as param (param.name)}
      <div class="px-5 py-2 border-t">
        <ParamValueInput
          {param}
          value={$argsStore[param.name ?? ""]}
          {metricsViewNames}
          metricsViewSpec={boundMetricsViewSpec(param, $argsStore)}
          valueOptions={isOrderByParam(param)
            ? orderByOptions(params, $argsStore)
            : undefined}
          onChange={(value) => setArg(param.name, value)}
        />
        {#if param.description}
          <div class="pt-1 text-[11px] text-fg-secondary">
            {param.description}
          </div>
        {/if}
        {#if hints[param.name ?? ""]}
          <div class="pt-1 text-[11px] text-fg-secondary">
            {hints[param.name ?? ""]}
          </div>
        {/if}
      </div>
    {/each}
  {/if}
</div>
