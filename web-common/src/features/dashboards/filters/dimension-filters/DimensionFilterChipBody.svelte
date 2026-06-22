<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let label: string;
  export let values: string[];
  export let matchedCount: number | undefined = undefined;
  export let search: string | undefined = undefined;
  export let loading: boolean | undefined = undefined;
  export let show = 1;
  export let smallChip = false;
  export let labelMaxWidth = "160px";
  export let valueMaxWidth = "320px";

  $: whatsLeft = values.length - show;
</script>

<div class="flex gap-x-2 items-center truncate">
  <span
    class="font-bold truncate"
    style:max-width={smallChip ? "150px" : labelMaxWidth}
  >
    {label}
  </span>

  {#if search}
    <span>{m.dashboards_filters_chip_contains()}</span>
    {#if loading}
      <Spinner status={EntityStatus.Running} size="10px" />
    {:else}
      <span class="italic">{search} ({matchedCount})</span>
    {/if}
  {:else if matchedCount !== undefined}
    <span>{m.dashboards_filters_chip_in_list()}</span>
    {#if loading}
      <Spinner status={EntityStatus.Running} size="10px" />
    {:else}
      <span class="italic"
        >{m.dashboards_filters_chip_matched_count({
          matched: matchedCount,
          total: values.length,
        })}</span
      >
    {/if}
  {:else}
    {#if !smallChip}
      {#each values.slice(0, show) as value (value)}
        <span class="truncate" style:max-width={valueMaxWidth}>
          {value}
        </span>
      {/each}
    {/if}

    {#if smallChip}
      <span class="italic">
        {m.dashboards_filters_chip_selected({ count: values.length })}
      </span>
    {:else if values.length > 1}
      <span class="italic flex-none">
        {m.dashboards_filters_chip_others({ count: whatsLeft })}
      </span>
    {/if}
  {/if}
</div>
