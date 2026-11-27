<script>
  import Spinner from "../../entity-management/Spinner.svelte";
  import { EntityStatus } from "../../entity-management/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import EmptyMeasureIcon from "./EmptyMeasureIcon.svelte";
  import EmptyTableIcon from "./EmptyTableIcon.svelte";

  export let isFetching = false;
  export let assembled = false;
  export let hasColumnAndNoMeasure = false;
  export let isEmbedded = false;
  /**
   * The height of the space the state is shown in, from a host that limits it.
   * @type {number | undefined}
   */
  export let availableHeight = undefined;

  // The height the tallest state needs: the spinner, its title and the help link.
  // A shorter space, such as a canvas table sized for a few rows,
  // gets one line of text instead.
  const FULL_STATE_HEIGHT = 168;

  $: compact =
    availableHeight !== undefined && availableHeight < FULL_STATE_HEIGHT;
</script>

<div class="flex flex-col items-center w-full h-full justify-center gap-y-6">
  {#if compact}
    <div class="compact">
      {#if isFetching}
        <Spinner size="14px" status={EntityStatus.Running} />
        <span>{m.dashboard_pivot_building_table()}</span>
      {:else if hasColumnAndNoMeasure}
        <span class="font-semibold text-fg-primary">
          {m.dashboard_pivot_keep_it_up()}
        </span>
        <span>{m.dashboard_pivot_add_measure()}</span>
      {:else if assembled}
        <span>{m.dashboard_pivot_no_data()}</span>
      {:else}
        <span class="font-semibold text-fg-primary">
          {m.dashboard_pivot_table_lonely()}
        </span>
        <span>{m.dashboard_pivot_give_data()}</span>
      {/if}
    </div>
  {:else if isFetching}
    <Spinner size="64px" status={EntityStatus.Running} />
    <div class="font-semibold text-fg-primary mt-1 text-lg">
      {m.dashboard_pivot_building_table()}
    </div>
    {#if !isEmbedded}
      <div class="text-fg-secondary">
        {m.dashboard_pivot_need_help_discord()}
        <a target="_blank" rel="noopener" href="https://discord.gg/2ubRfjC7Rh">
          <!-- i18n-ignore: standalone product name -->
          Discord</a
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
          >{m.common_docs()}</a
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
          >{m.common_docs()}</a
        >.
      </div>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .compact {
    @apply flex flex-wrap items-center justify-center gap-x-2 px-2;
    @apply text-sm text-center text-fg-secondary;
  }
</style>
