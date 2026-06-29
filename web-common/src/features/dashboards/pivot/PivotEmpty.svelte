<script>
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import EmptyMeasureIcon from "./EmptyMeasureIcon.svelte";
  import EmptyTableIcon from "./EmptyTableIcon.svelte";

  export let isFetching = false;
  export let assembled = false;
  export let hasColumnAndNoMeasure = false;
  export let isEmbedded = false;
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-6">
  {#if isFetching}
    <Spinner size="64px" status={EntityStatus.Running} />
    <div class="font-semibold text-fg-primary mt-1 text-lg">
      {m.dashboard_pivot_building_table()}
    </div>
    {#if !isEmbedded}
      <div class="text-fg-secondary">
        {m.dashboard_pivot_need_help_discord()}
        <a target="_blank" rel="noopener" href="https://discord.gg/2ubRfjC7Rh"
          >Discord</a
        >
      </div>
    {/if}
  {:else if hasColumnAndNoMeasure}
    <EmptyMeasureIcon />
    <div class="flex flex-col items-center gap-y-2">
      <div class="font-semibold text-fg-primary mt-1 text-lg">
        {m.dashboard_pivot_keep_it_up()}
      </div>
      <div class="text-fg-secondary text-base">
        {m.dashboard_pivot_add_measure()}
      </div>
    </div>
    {#if !isEmbedded}
      <div class="text-fg-secondary">
        {m.dashboard_pivot_learn_more()}
        <a
          target="_blank"
          rel="noopener"
          href="https://docs.rilldata.com/guide/dashboards/explore/pivot"
          >docs</a
        >.
      </div>
    {/if}
  {:else if assembled}
    <EmptyTableIcon />
    <div class="text-fg-secondary text-base">
      {m.dashboard_pivot_no_data()}
    </div>
  {:else}
    <EmptyTableIcon />
    <div class="flex flex-col items-center gap-y-2">
      <div class="font-semibold text-fg-primary mt-1 text-lg">
        {m.dashboard_pivot_table_lonely()}
      </div>
      <div class="text-fg-secondary text-base">
        {m.dashboard_pivot_give_data()}
      </div>
    </div>
    {#if !isEmbedded}
      <div class="text-fg-secondary">
        {m.dashboard_pivot_learn_more()}
        <a
          target="_blank"
          href="https://docs.rilldata.com/guide/dashboards/explore/pivot"
          >docs</a
        >.
      </div>
    {/if}
  {/if}
</div>
