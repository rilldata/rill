<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { type CanvasComponent } from "@rilldata/web-common/features/canvas/components/types";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import { isString } from "@rilldata/web-common/features/canvas/util";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let component: CanvasComponent<any>;
  export let key: string;
  export let inputParam: ComponentInputParam;

  $: ({ instanceId } = $runtime);

  $: spec = component.specStore;

  $: metricsViewsQuery = useFilteredResources(
    instanceId,
    ResourceKind.MetricsView,
  );
  $: metricsViews = $metricsViewsQuery?.data ?? [];

  $: metricsViewNames = metricsViews
    .map((view) => view.meta?.name?.name)
    .filter(isString);

  $: metricsView =
    "metrics_view" in $spec ? $spec.metrics_view : metricsViewNames[0];
</script>

<Input
  hint="View documentation"
  link="https://docs.rilldata.com/reference/project-files/metrics-view"
  label={inputParam.label}
  capitalizeLabel={false}
  bind:value={metricsView}
  sameWidth
  options={metricsViewNames.map((name) => ({
    label: name,
    value: name,
  }))}
  onBlur={() => {
    component.updateProperty(key, metricsView);
  }}
  onChange={() => {
    component.updateProperty(key, metricsView);
  }}
/>
