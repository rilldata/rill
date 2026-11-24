<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { CalendarIcon } from "lucide-svelte";
  import { DateTime } from "luxon";
  import {
    DEFAULT_TOKEN_EXPIRATION,
    formatTokenDate,
    type TokenExpiration,
  } from "./personal-access-tokens";

  let {
    id,
    expiration = $bindable(DEFAULT_TOKEN_EXPIRATION),
    customDate = $bindable(null),
  }: {
    id: string;
    expiration?: TokenExpiration;
    // The day picked for the "custom" expiration, at the start of the day.
    customDate?: DateTime | null;
  } = $props();

  const options: { value: TokenExpiration; label: string }[] = [
    { value: "7d", label: m.access_tokens_expiration_days({ count: 7 }) },
    { value: "30d", label: m.access_tokens_expiration_days({ count: 30 }) },
    { value: "90d", label: m.access_tokens_expiration_days({ count: 90 }) },
    { value: "1y", label: m.access_tokens_expiration_one_year() },
    { value: "custom", label: m.access_tokens_expiration_custom() },
    { value: "none", label: m.access_tokens_expiration_none() },
  ];

  const minDate = DateTime.now().plus({ days: 1 }).startOf("day");
  const maxDate = DateTime.now().plus({ years: 1 }).startOf("day");

  let calendarOpen = $state(false);
  let customDateLabel = $derived(customDate ? formatTokenDate(customDate) : "");

  function onChange(value: string) {
    expiration = value as TokenExpiration;
    if (expiration === "custom" && !customDate) {
      customDate = DateTime.now().plus({ days: 30 }).startOf("day");
    }
  }
</script>

<div class="flex flex-col gap-y-2">
  <Select
    {id}
    label={m.access_tokens_expiration_label()}
    value={expiration}
    {options}
    {onChange}
    sameWidth
    full
  />
  {#if expiration === "custom" && customDate}
    <Popover bind:open={calendarOpen}>
      <PopoverTrigger>
        {#snippet child({ props })}
          <Button
            {...props}
            type="secondary"
            label={m.access_tokens_expiration_pick_date()}
            fit
          >
            <CalendarIcon size="14px" />
            {customDateLabel}
          </Button>
        {/snippet}
      </PopoverTrigger>
      <PopoverContent align="start" class="p-0" strategy="fixed">
        <Calendar
          selection={customDate}
          singleDaySelection
          {minDate}
          {maxDate}
          firstVisibleMonth={customDate}
          onSelectDay={(date) => {
            customDate = date;
            calendarOpen = false;
          }}
        />
      </PopoverContent>
    </Popover>
  {/if}
</div>
