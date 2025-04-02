<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import { isString } from "../../workspaces/visual-util";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { AllKeys } from "node_modules/sveltekit-superforms/dist/utils";
  import type { ComponentSpec } from "../components/types";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";

  export let component: BaseCanvasComponent;
  export let key: AllKeys<ComponentSpec>;
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
  size="sm"
  labelGap={2}
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
