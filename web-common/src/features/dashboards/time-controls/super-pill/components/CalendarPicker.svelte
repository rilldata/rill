<script lang="ts">
  import { Interval } from "luxon";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Calendar2 from "@rilldata/web-common/components/date-picker/Calendar2.svelte";
  import { DateTime } from "luxon";

  const formatsWithoutYear = [
    "M/d", // 7/4 or 07/04
    "MMM d", // Jul 4 or July 04
    "MMMM d", // July 4 or July 04
  ];

  const formatsWithYear = [
    "M/d/yy", // 7/4/21 or 07/04/21
    "D", // 7/4/2021 or 07/04/2021
    "DDD", // July 4, 2021 or July 04, 2021
    "MMM d, yyyy", // July 4, 2021 or July 04, 2021
    "MMM d yyyy", // July 4, 2021 or July 04, 2021
    "MMMM d yyyy", // July 4 2021 or July 04 2021
  ];

  const formats = [...formatsWithoutYear, ...formatsWithYear];

  export let interval: Interval<true>;
  export let zone: string;

  let open = false;
  let selectingStart = true;
  let startInput: HTMLInputElement;
  let endInput: HTMLInputElement;

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
        break;
      }
    }

    if (!date.isValid) return;

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

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger aria-label="Select a time range" asChild let:builder>
    <button
      use:builder.action
      {...builder}
      class="flex-none flex items-center justify-center pb-[1.5px] hover:bg-gray-200"
    >
      <Calendar size="16px" />
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-72">
    <Calendar2 bind:interval bind:selectingStart bind:firstVisibleMonth />
    <DropdownMenu.Separator />
    <div class="flex flex-col gap-y-2 px-2 pt-1 pb-2">
      <label for="start-date">
        Start Date <span class="secondary">(Inclusive)</span>
      </label>
      <div class="flex gap-x-1">
        <input
          id="start-date"
          type="text"
          bind:this={startInput}
          class:active={selectingStart}
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
      <label for="start-date">
        End Date <span class="secondary">(Exclusive)</span>
      </label>
      <div class="flex gap-x-1 w-full">
        <input
          id="end-date"
          type="text"
          bind:this={endInput}
          on:blur={validateInput}
          class:active={!selectingStart}
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
    @apply font-semibold;
  }
</style>
