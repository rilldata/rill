<script lang="ts">
  import { Chip } from "../../../../components/chip";
  import {
    defaultChipColors,
    excludeChipColors,
  } from "../../../../components/chip/chip-types";

  export let label: string;
  export let values: string[];
  export let isInclude: boolean;

  const show = 1;
  const labelMaxWidth = "160px";
  const valueMaxWidth = "320px";

  $: visibleValues = values.slice(0, show);
  $: whatsLeft = values.length - show;
  $: effectiveLabel = isInclude ? label : `Exclude ${label}`;
  $: colors = isInclude ? defaultChipColors : excludeChipColors;
</script>

<Chip {...colors} label={effectiveLabel} outline readOnly>
  <svelte:fragment slot="body">
    <div class="flex gap-x-2 px-2">
      <div
        class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
        style:max-width={labelMaxWidth}
      >
        {effectiveLabel}
      </div>
      <div class="flex flex-wrap flex-row items-center gap-y-1 gap-x-2">
        {#each visibleValues as value}
          <div
            class="text-ellipsis overflow-hidden whitespace-nowrap"
            style:max-width={valueMaxWidth}
          >
            {value}
          </div>
        {/each}
        {#if values.length > 1}
          <div class="italic">
            +{whatsLeft} other{#if whatsLeft !== 1}s{/if}
          </div>
        {/if}
      </div>
    </div>
  </svelte:fragment>
</Chip>
