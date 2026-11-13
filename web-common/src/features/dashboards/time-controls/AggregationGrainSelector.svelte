<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { translateV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    timeFilterManager,
    type = "default",
  }: {
    timeFilterManager: TimeFilterManager;
    // "tdd" renders the trigger as a pill labelled "Time Day" for the time dimension details header.
    type?: "tdd" | "default";
  } = $props();

  let open = $state(false);

  let { urlTimeGrain, timeGrain, aggregationOptions } =
    $derived(timeFilterManager);

  let isTdd = $derived(type === "tdd");
  let defaultOptionSelected = $derived(!urlTimeGrain);
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <button
        {...props}
        aria-label={m.dashboard_select_aggregation_grain_aria()}
        class={[
          "flex gap-x-1 items-center",
          isTdd ? "tdd" : "text-fg-muted hover:text-fg-accent",
        ]}
      >
        {#if isTdd}
          <b>{m.time_grain_time()}</b>
          <span class="capitalize">{translateV1TimeGrain(timeGrain)}</span>
        {:else}
          {m.time_grain_by()}
          <b>
            {translateV1TimeGrain(timeGrain)}
          </b>
        {/if}
        {#if defaultOptionSelected}
          <i class={{ "text-fg-muted": isTdd }}>
            {m.time_grain_adaptive_hint()}
          </i>
        {/if}
        <span
          class={[
            "transition-transform",
            { "-rotate-90": open && !isTdd, "-rotate-180": open && isTdd },
          ]}
        >
          <CaretDownIcon />
        </span>
      </button>
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class={isTdd ? "min-w-52" : "w-48"}>
    <DropdownMenu.CheckboxItem
      checkRight
      checked={defaultOptionSelected}
      class="text-xs cursor-pointer"
      onclick={() => timeFilterManager.onSelectGrain(undefined)}
    >
      <div class="flex flex-col">
        <span>{m.time_grain_adaptive()}</span>
        <span class="text-fg-muted">{m.time_grain_adaptive_description()}</span>
      </div>
    </DropdownMenu.CheckboxItem>
    <DropdownMenu.Separator />
    {#each aggregationOptions as option (option)}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={option === urlTimeGrain}
        class={["text-xs cursor-pointer", { capitalize: isTdd }]}
        onclick={() => timeFilterManager.onSelectGrain(option)}
      >
        {translateV1TimeGrain(option)}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  .tdd {
    @apply border h-7 rounded-full px-2 pl-2.5 bg-surface-background text-fg-primary;
  }

  .tdd:global([data-state="open"]) {
    @apply border-gray-400;
  }
</style>
