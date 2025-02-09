<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import MultiIconSelector from "@rilldata/web-common/components/forms/MultiIconSelector.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import DeltaPercentage from "@rilldata/web-common/components/icons/DeltaPercentage.svelte";
  import { defaultComparisonOptions } from "@rilldata/web-common/features/canvas/components/kpi";
  import { type ComponentComparisonOptions } from "@rilldata/web-common/features/canvas/components/types";
  import type { ComponentType, SvelteComponent } from "svelte";
  import CounterClockWiseClock from "svelte-radix/CounterClockwiseClock.svelte";

  export let key: string;
  export let label: string;
  export let options: string[] | undefined;
  export let onChange: (options: string[]) => void;

  const comparisonOptions: {
    id: ComponentComparisonOptions;
    Icon: ComponentType<SvelteComponent>;
  }[] = [
    { id: "delta", Icon: Delta },
    { id: "percent_change", Icon: DeltaPercentage },
    { id: "previous", Icon: CounterClockWiseClock },
  ];
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex justify-between">
    <InputLabel small {label} id={key} faint={!options} />
    <Switch
      checked={!!options?.length}
      on:click={() => {
        onChange(options?.length ? [] : defaultComparisonOptions);
      }}
      small
    />
  </div>

  {#if options}
    <MultiIconSelector
      small
      expand
      fields={comparisonOptions}
      selected={options || defaultComparisonOptions}
      onChange={(option) => onChange(option)}
    />
  {/if}
</div>
