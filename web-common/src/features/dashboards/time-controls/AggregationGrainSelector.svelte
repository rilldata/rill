<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { translateV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";

  // The "by Day" aggregation grain dropdown shared by the explore chart toolbar,
  // the canvas filter bar and the canvas widget inspector.
  let {
    timeFilterManager,
    defaultOption = undefined,
  }: {
    timeFilterManager: TimeFilterManager;
    // Listed above the grains: "Adaptive" in explore and the canvas filter bar,
    // "Inherit from canvas" in the widget inspector. `hint` shows next to the trigger while selected.
    defaultOption?: {
      label: string;
      description?: string;
      hint?: string;
    };
  } = $props();

  let { urlTimeGrain, timeGrain, aggregationOptions } =
    $derived(timeFilterManager);

  let open = $state(false);
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <button
        {...props}
        aria-label={m.dashboard_select_aggregation_grain_aria()}
        class="flex gap-x-1 items-center text-fg-muted hover:text-fg-accent"
      >
        {m.explore_by_grain_prefix()}
        <b>
          {translateV1TimeGrain(timeGrain)}
        </b>
        {#if !urlTimeGrain && defaultOption?.hint}
          <i>{defaultOption.hint}</i>
        {/if}
        <span class={["transition-transform", { "-rotate-90": open }]}>
          <CaretDownIcon />
        </span>
      </button>
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-48">
    {#if defaultOption}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={!urlTimeGrain}
        class="text-xs cursor-pointer"
        onclick={() => timeFilterManager.onSelectGrain(undefined)}
      >
        <div class="flex flex-col">
          <span>{defaultOption.label}</span>
          {#if defaultOption.description}
            <span class="text-fg-muted">{defaultOption.description}</span>
          {/if}
        </div>
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.Separator />
    {/if}
    {#each aggregationOptions as option (option)}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={option === urlTimeGrain}
        class="text-xs cursor-pointer"
        onclick={() => timeFilterManager.onSelectGrain(option)}
      >
        {translateV1TimeGrain(option)}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
