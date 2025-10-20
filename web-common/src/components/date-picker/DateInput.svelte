<script lang="ts">
  import { DateTime } from "luxon";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import AlertTriangle from "../icons/AlertTriangle.svelte";

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
    "MMM d, yyyy", // Jul 4, 2021
    "MMM d yyyy", // Jul 4 2021
    "MMMM d yyyy", // July 4 2021 or July 04 2021
    "yyyy-M-d", // July 4, 2021 or July 04, 2021
    "M-d-yyyy", // July 4, 2021 or July 04, 2021
    "M-d-yy", // 7-4-21 or 07-04-21
    "d MMMM yyyy", // 4 July 2021 or 04 July 2021
    "d MMM yyyy", // 05 Jul 2021 or 5 Jul 2021
    "d MMMM, yyyy", // 4 July 2021 or 04 July 2021
    "d MMM, yyyy", // 05 Jul 2021 or 5 Jul 2021
    "d.M.yyyy", // 4.7.2021 or 04.07.2021
  ];

  const formats = [...formatsWithoutYear, ...formatsWithYear];

  enum ErrorType {
    INVALID = "invalid",
    OUT_OF_RANGE = "out-of-range",
  }

  export let date: DateTime;
  export let zone: string;
  export let boundary: "start" | "end";
  export let currentYear: number;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let onValidDateInput: (
    date: DateTime,
    boundary: "start" | "end",
  ) => void;
  export let onFocus: () => void;

  let errorType: ErrorType | null = null;
  let inputIsFocused = false;
  let dateString = date.toLocaleString({
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  $: id = boundary + "-date";
  $: label = boundary + " date";

  $: if (!inputIsFocused) {
    dateString = date.toLocaleString({
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  }

  $: if (
    (minDate && date < minDate.startOf("day")) ||
    (maxDate && date >= maxDate)
  ) {
    errorType = ErrorType.OUT_OF_RANGE;
  } else {
    errorType = null;
  }

  function convertToDateTime(dateString: string): DateTime<true> | undefined {
    let potentialDate: DateTime | undefined = undefined;
    let format: string | null = null;

    for (const potentialFormat of formats) {
      potentialDate = DateTime.fromFormat(dateString, potentialFormat, {
        zone,
      });

      if (potentialDate.isValid) {
        format = potentialFormat;

        break;
      }
    }

    if (
      potentialDate &&
      potentialDate.year !== currentYear &&
      format &&
      formatsWithoutYear.includes(format)
    ) {
      potentialDate = potentialDate.set({ year: currentYear });
    }

    if (potentialDate?.isValid) {
      return potentialDate;
    } else {
      return undefined;
    }
  }

  function processInput(value: string) {
    let potentialDate = convertToDateTime(value);

    if (potentialDate) {
      if (boundary === "end") {
        potentialDate = potentialDate?.plus({ day: 1 }).startOf("day");
      }

      onValidDateInput(potentialDate, boundary);
    } else {
      errorType = ErrorType.INVALID;
      return;
    }
  }

  function resetDate() {
    if (boundary === "start") {
      date = minDate ?? DateTime.now().startOf("day");
      onValidDateInput(date, boundary);
    } else {
      date = maxDate ?? DateTime.now().plus({ day: 1 }).startOf("day");
      onValidDateInput(date, boundary);
    }
  }
</script>

<div class="flex flex-col gap-y-1 w-full">
  <label class="capitalize flex items-center" for={id}>
    {label}
  </label>

  <div class="input-wrapper" class:error={errorType === ErrorType.INVALID}>
    <input
      tabindex="0"
      {id}
      class="size-full bg-transparent"
      aria-label={label}
      type="text"
      bind:value={dateString}
      on:focus={() => {
        onFocus();
        inputIsFocused = true;
      }}
      on:blur={() => {
        processInput(dateString);
        inputIsFocused = false;
      }}
    />
    {#if errorType === ErrorType.OUT_OF_RANGE || (errorType && !inputIsFocused)}
      <Tooltip.Root portal="body">
        <Tooltip.Trigger asChild let:builder>
          <button use:builder.action {...builder} on:click={resetDate}>
            <AlertTriangle
              className="size-4 text-{errorType === ErrorType.INVALID
                ? 'red'
                : 'yellow'}-500"
            />
          </button>
        </Tooltip.Trigger>
        <Tooltip.Content
          side="top"
          sideOffset={10}
          class="bg-gray-700 text-surface shadow-md"
        >
          {#if errorType === ErrorType.OUT_OF_RANGE}
            Date is out of range. Click to reset.
          {:else}
            Date is invalid. Click to reset.
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}
  </div>
</div>

<style lang="postcss">
  .input-wrapper {
    @apply h-8 px-2 w-full rounded-md border border-gray-300 flex bg-surface;
    @apply items-center justify-between;
  }

  .input-wrapper:focus-within {
    outline: none;
    @apply border-primary-600;
  }

  input:focus {
    outline: none;
  }

  .input-wrapper.error:not(:focus-within) {
    @apply border-destructive text-destructive;
  }

  label {
    @apply font-semibold flex gap-x-1;
  }
</style>
