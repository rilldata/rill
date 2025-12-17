<script lang="ts">
  import type { GaugeComponent } from ".";
  import ComponentHeader from "../../ComponentHeader.svelte";
  import { validateGaugeSchema } from "./selector";
  import GaugeProvider from "./GaugeProvider.svelte";
  import { getCanvasStore } from "../../state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let component: GaugeComponent;

  $: ({
    specStore,
    timeAndFilterStore,
    parent: { name: canvasName },
    visible,
  } = component);
  $: gaugeProperties = $specStore;
  $: ctx = getCanvasStore(canvasName, $runtime.instanceId);
  $: schema = validateGaugeSchema(ctx, gaugeProperties);

  $: ({ title, description, show_description_as_tooltip } = gaugeProperties);

  $: filters = {
    time_filters: gaugeProperties.time_filters,
    dimension_filters: gaugeProperties.dimension_filters,
  };
</script>

<ComponentHeader
  {component}
  {title}
  {description}
  showDescriptionAsTooltip={show_description_as_tooltip}
  {filters}
/>

{#if schema.isValid}
  <div class="gauge-component-wrapper h-full w-full p-4">
    <GaugeProvider
      spec={gaugeProperties}
      {timeAndFilterStore}
      {canvasName}
      {visible}
    />
  </div>
{/if}

<style lang="postcss">
  .gauge-component-wrapper {
    @apply flex items-center justify-center;
    min-height: 200px;
  }
</style>

