<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  const { unnestAndFilter } = featureFlags;

  export let label: string;
  export let values: string[];
  export let matchedCount: number | undefined = undefined;
  export let search: string | undefined = undefined;
  export let loading: boolean | undefined = undefined;
  export let andMode: boolean = false;
  export let isUnnest: boolean = false;
  export let show = 1;
  export let smallChip = false;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";

  $: whatsLeft = values.length - show;
  $: operator =
    $unnestAndFilter && isUnnest
      ? andMode === true
        ? "AND"
        : "OR"
      : undefined;
</script>

<div class="flex gap-x-2 items-center truncate">
  <span
    class="font-bold truncate"
    style:max-width={smallChip ? "150px" : labelMaxWidth}
  >
    {label}
  </span>

  {#if search}
    <span>Contains</span>
    {#if loading}
      <Spinner status={EntityStatus.Running} size="10px" />
    {:else}
      <span class="italic">{search} ({matchedCount})</span>
    {/if}
  {:else if matchedCount !== undefined}
    <span>In list</span>
    {#if loading}
      <Spinner status={EntityStatus.Running} size="10px" />
    {:else}
      <span class="italic">({matchedCount} of {values.length})</span>
    {/if}
  {:else}
    {#if operator}
      <span
        class="text-fg-secondary text-[10px] uppercase font-semibold flex-none"
        >{operator}</span
      >
    {/if}
    {#if !smallChip}
      {#each values.slice(0, show) as value (value)}
        <span class="truncate" style:max-width={valueMaxWidth}>
          {value}
        </span>
      {/each}
    {/if}

    {#if smallChip}
      <span class="italic">
        {values.length} selected
      </span>
    {:else if values.length > 1}
      <span class="italic flex-none">
        +{whatsLeft} other{#if whatsLeft !== 1}s{/if}
      </span>
    {/if}
  {/if}
</div>
