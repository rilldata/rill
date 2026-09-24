<script lang="ts">
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useDefaultMetrics } from "@rilldata/web-common/features/canvas/selector";
  import {
    buildDefaultArgs,
    getDeclaredParams,
    isOrderByParam,
    orderByOptions,
    reconcileOrderByArg,
  } from "@rilldata/web-common/features/custom-viz/params";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    type V1ComponentParam,
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

  // Seed default test bindings once per param-shape change, for params without a value.
  $: seedDefaults(params, defaultMetrics);

  function seedDefaults(
    params: V1ComponentParam[],
    defaultMetrics:
      | { metricsViewName: string; metricsViewSpec: any }
      | null
      | undefined,
  ) {
    if (!params.length || !defaultMetrics) return;
    const defaults = buildDefaultArgs(
      params,
      defaultMetrics.metricsViewName,
      defaultMetrics.metricsViewSpec,
    );
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
      </div>
    {/each}
  {/if}
</div>
