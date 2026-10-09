<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { translateV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  // The "by Day" aggregation grain dropdown shared by the explore chart toolbar,
  // the canvas filter bar and the canvas widget inspector.
  let {
    activeTimeGrain,
    options,
    onSelect,
    inheritOption = undefined,
  }: {
    activeTimeGrain: V1TimeGrain;
    options: V1TimeGrain[];
    onSelect: (timeGrain: V1TimeGrain) => void;
    // Offered by the widget inspector so a widget can drop its own grain again.
    inheritOption?: { label: string; selected: boolean; onSelect: () => void };
  } = $props();

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
          {translateV1TimeGrain(activeTimeGrain)}
        </b>
        <span class={["transition-transform", { "-rotate-90": open }]}>
          <CaretDownIcon />
        </span>
      </button>
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-48">
    {#if inheritOption}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={inheritOption.selected}
        class="text-xs cursor-pointer"
        onclick={inheritOption.onSelect}
      >
        {inheritOption.label}
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.Separator />
    {/if}
    {#each options as option (option)}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={!inheritOption?.selected && option === activeTimeGrain}
        class="text-xs cursor-pointer"
        onclick={() => onSelect(option)}
      >
        {translateV1TimeGrain(option)}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
