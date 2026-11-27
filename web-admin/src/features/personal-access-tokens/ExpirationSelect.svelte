<script lang="ts">
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { CalendarIcon } from "lucide-svelte";
  import { DateTime } from "luxon";
  import {
    DEFAULT_TOKEN_EXPIRATION,
    formatTokenDate,
    getTokenExpiryDate,
    type TokenExpiration,
  } from "./personal-access-tokens";

  let {
    id,
    expiration = $bindable(DEFAULT_TOKEN_EXPIRATION),
    customDate = $bindable(""),
  }: {
    id: string;
    expiration?: TokenExpiration;
    // The day picked for the "custom" expiration, as yyyy-MM-dd.
    customDate?: string;
  } = $props();

  const presets: {
    value: Exclude<TokenExpiration, "custom">;
    label: string;
  }[] = [
    { value: "7d", label: m.access_tokens_expiration_days({ count: 7 }) },
    { value: "30d", label: m.access_tokens_expiration_days({ count: 30 }) },
    { value: "90d", label: m.access_tokens_expiration_days({ count: 90 }) },
    { value: "1y", label: m.access_tokens_expiration_one_year() },
    { value: "none", label: m.access_tokens_expiration_none() },
  ];

  const minDate = DateTime.now().plus({ days: 1 }).startOf("day");
  const maxDate = DateTime.now().plus({ years: 1 }).startOf("day");

  let open = $state(false);
  // As in the dashboard time range picker, "Custom date" opens a calendar to the right of the options.
  let showCalendar = $state(false);

  let expiresOn = $derived(getTokenExpiryDate(expiration, customDate));
  let selectedLabel = $derived(
    expiration === "custom" && expiresOn
      ? formatTokenDate(expiresOn)
      : (presets.find((p) => p.value === expiration)?.label ?? ""),
  );
  // The calendar opens on the current expiry, or a month out for tokens that never expire.
  let calendarDay = $derived(
    (expiresOn ?? DateTime.now().plus({ days: 30 })).startOf("day"),
  );
</script>

<div class="flex flex-col gap-y-2">
  <label for={id} class="text-sm font-medium text-fg-primary">
    {m.access_tokens_expiration_label()}
  </label>
  <Popover.Root
    bind:open
    onOpenChange={(isOpen) => {
      if (isOpen) showCalendar = expiration === "custom";
    }}
  >
    <Popover.Trigger
      {id}
      class="flex h-8 w-full items-center justify-between gap-x-2 rounded-[2px] border bg-input px-3 text-xs text-fg-primary focus:outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
    >
      <span class="truncate">{selectedLabel}</span>
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon size="12px" className="fill-fg-secondary" />
      </span>
    </Popover.Trigger>
    <Popover.Content
      align="start"
      padding="0"
      strategy="fixed"
      class="flex w-fit overflow-hidden"
    >
      <div class="flex w-44 flex-none flex-col p-1">
        {#each presets as preset (preset.value)}
          <button
            type="button"
            role="menuitem"
            class="option"
            class:font-semibold={expiration === preset.value}
            onclick={() => {
              expiration = preset.value;
              open = false;
            }}
          >
            {preset.label}
          </button>
        {/each}
        <div class="my-1 h-px w-full bg-border"></div>
        <button
          type="button"
          role="menuitem"
          class="option gap-x-1.5"
          class:font-semibold={expiration === "custom"}
          aria-expanded={showCalendar}
          onclick={() => (showCalendar = !showCalendar)}
        >
          <CalendarIcon size="14px" />
          <span class="mr-auto">{m.access_tokens_expiration_custom()}</span>
          <CaretDownIcon size="12px" className="-rotate-90 fill-fg-secondary" />
        </button>
      </div>
      {#if showCalendar}
        <div class="w-64 border-l p-3">
          <Calendar
            selection={calendarDay}
            singleDaySelection
            {minDate}
            {maxDate}
            firstVisibleMonth={calendarDay}
            onSelectDay={(date) => {
              customDate = date.toISODate();
              expiration = "custom";
              open = false;
            }}
          />
        </div>
      {/if}
    </Popover.Content>
  </Popover.Root>

  {#if expiresOn}
    <p class="text-xs text-fg-secondary">
      {m.access_tokens_expires_on({ date: formatTokenDate(expiresOn) })}
    </p>
  {:else}
    <Callout level="warning">
      <span class="text-xs">{m.access_tokens_no_expiration_warning()}</span>
    </Callout>
  {/if}
</div>

<style lang="postcss">
  .option {
    @apply flex h-7 w-full select-none items-center rounded-sm px-2 text-left text-xs text-fg-primary;
  }

  .option:hover {
    @apply bg-popover-accent;
  }
</style>
