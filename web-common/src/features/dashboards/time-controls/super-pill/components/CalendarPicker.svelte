<script context="module" lang="ts">
  import { writable } from "svelte/store";
  import { Interval } from "luxon";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CalendarIcon from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import { DateTime } from "luxon";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export const open = writable(false);

  const formatsWithoutYear = [
    "M/d", // 7/4 or 07/04
    "MMMM d", // July 4 or July 04
    "MMM d", // Jul 4 or Jul 04
    "M-d", // 7-4 or 07-04
    "d MMMM", // 4 July or 04 July
    "d MMM", // 4 Jul or 04 Jul
    "d.M", // 4.7 or 04.07
  ];

  const formatsWithYear = [
    "M/d/yy", // 7/4/21 or 07/04/21
    "D", // 7/4/2021 or 07/04/2021
    "DDD", // July 4, 2021 or July 04, 2021
    "MMM d, yyyy", // July 4, 2021 or July 04, 2021
    "MMM d yyyy", // July 4, 2021 or July 04, 2021
    "MMMM d yyyy", // July 4 2021 or July 04 2021
    "yyyy-M-d", // July 4, 2021 or July 04, 2021
    "M-d-yyyy", // July 4, 2021 or July 04, 2021
    "d MMMM yyyy", // 4 July 2021 or 04 July 2021
    "d MMM yyyy", // 05 Jul 2021 or 5 Jul 2021
    "d MMMM, yyyy", // 4 July 2021 or 04 July 2021
    "d MMM, yyyy", // 05 Jul 2021 or 5 Jul 2021
    "d.M.yyyy", // 4.7.2021 or 04.07.2021
  ];

  const formats = [...formatsWithoutYear, ...formatsWithYear];
</script>

<script lang="ts">
  export let interval: Interval<true>;
  export let zone: string;
  export let applyRange: (range: Interval<true>) => void;

  let selectingStart = true;
  let startInput: HTMLInputElement;
  let endInput: HTMLInputElement;
  let displayError = false;

  let firstVisibleMonth: DateTime<true>;

  function validateInput(
    e: FocusEvent & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const dateString = e.currentTarget.value;

    let date: DateTime = DateTime.invalid("invalid");

    let format: string | null = null;

    for (const potentialFormat of formats) {
      date = DateTime.fromFormat(dateString, potentialFormat, {
        zone,
      });

      if (date.isValid) {
        format = potentialFormat;
        displayError = false;
        break;
      }
    }

    if (!date.isValid) {
      displayError = true;
      return;
    }

    if (
      date.year !== firstVisibleMonth.year &&
      format &&
      formatsWithoutYear.includes(format)
    ) {
      date = date.set({ year: firstVisibleMonth.year });
    }

    const newInterval = interval.set({
      [selectingStart ? "start" : "end"]: date,
    });

    if (newInterval.isValid) {
      interval = newInterval;
    } else {
      interval = Interval.fromDateTimes(
        date,
        date.plus({ day: 1 }),
      ) as Interval<true>;
    }

    firstVisibleMonth = interval.start;
  }
</script>

<DropdownMenu.Root bind:open={$open}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      aria-label="Select a custom time range"
      class="flex-none flex items-center justify-center pb-[1.5px] hover:bg-gray-200"
    >
      <CalendarIcon size="16px" />
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-72">
    <Calendar bind:interval bind:selectingStart bind:firstVisibleMonth />
    <DropdownMenu.Separator />
    <div class="flex flex-col gap-y-2 px-2 pt-1 pb-2">
      <label for="start-date" class:error={selectingStart && displayError}>
        Start Date <span class="secondary">(Inclusive)</span>
      </label>
      <div class="flex gap-x-1">
        <input
          id="start-date"
          aria-label="Start date"
          type="text"
          bind:this={startInput}
          class:active={selectingStart}
          class:error={displayError}
          on:click={() => {
            selectingStart = true;
          }}
          on:keydown={(e) => {
            if (e.key === "Enter") {
              e.currentTarget.blur();
            }
          }}
          on:blur={validateInput}
          value={interval?.start?.toLocaleString({
            month: "long",
            day: "2-digit",
            year: "numeric",
          })}
        />

        <!-- <input type="text" value={interval.start.toFormat("hh:mm a")} /> -->
      </div>
      <label for="start-date" class:error={!selectingStart && displayError}>
        End Date <span class="secondary">(Exclusive)</span>
      </label>
      <div class="flex gap-x-1 w-full">
        <input
          id="end-date"
          aria-label="End date"
          type="text"
          bind:this={endInput}
          on:blur={validateInput}
          class:active={!selectingStart}
          class:error={displayError}
          on:click={() => {
            selectingStart = false;
          }}
          on:keydown={(e) => {
            if (e.key === "Enter") {
              e.currentTarget.blur();
            }
          }}
          value={interval?.end?.toLocaleString({
            month: "long",
            day: "2-digit",
            year: "numeric",
          })}
        />
        <!-- <input type="text" value={interval?.end?.toFormat("hh:mm a")} /> -->
      </div>
    </div>
    <div class="flex justify-end w-full py-1 px-2">
      <Button
        fit
        compact
        on:click={() => {
          applyRange(interval);
          open.set(false);
        }}
      >
        <span class="px-2 w-fit">Apply</span>
      </Button>
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #bebebe;
    border-radius: 0.25rem;
  }

  input.active {
    outline: none;
    border-color: #007bff;
  }
  label {
    @apply font-semibold flex gap-x-1;
  }
  input.error.active {
    @apply border-destructive text-destructive;
  }

  /* label.error {
    @apply text-destructive;
  } */
</style>
