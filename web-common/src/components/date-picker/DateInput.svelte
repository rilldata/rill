<script lang="ts">
  import {
    CalendarDaysIcon,
    CalendarIcon,
    CalendarX2Icon,
  } from "lucide-svelte";
  import { DateTime } from "luxon";

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

  export let selectingStart: boolean;
  export let date: DateTime;
  export let zone: string;
  export let label: "from" | "to";
  export let currentYear: number;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let onValidDateInput: (date: DateTime) => void;

  let initialValue: string | null = null;
  let displayError: boolean;

  $: id = label + "-input";

  function validateInput(
    e: FocusEvent & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const changed = e.currentTarget.value !== initialValue;
    if (!changed) return;

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

    if (minDate && date < minDate) {
      displayError = true;
      return;
    }

    if (maxDate && date > maxDate) {
      displayError = true;
      return;
    }

    if (
      date.year !== currentYear &&
      format &&
      formatsWithoutYear.includes(format)
    ) {
      date = date.set({ year: currentYear });
    }

    onValidDateInput(date);
  }
</script>

<div class="flex flex-col gap-y-1 w-full">
  <label
    class="capitalize"
    for={id}
    class:error={selectingStart && displayError}
  >
    {label}
  </label>

  <div class="input-wrapper">
    <input
      tabindex="0"
      {id}
      aria-label={label}
      type="text"
      class:active={(label === "from") === selectingStart}
      class:error={displayError}
      value={date.toLocaleString({
        month: "short",
        day: "numeric",
        year: "numeric",
      })}
      on:click={(e) => {
        selectingStart = label === "to";
        initialValue = e.currentTarget.value;
      }}
      on:keydown={({ currentTarget, key }) => {
        if (key === "Enter") {
          currentTarget.blur();
        }
      }}
      on:blur={validateInput}
    />

    <!-- <button
      class="bg-gray-200/50 hover:bg-gray-200/70 active:bg-gray-200 aspect-square h-full border-l grid place-content-center"
    >
      <CalendarDaysIcon size="15px" />
    </button> -->
  </div>
</div>

<style lang="postcss">
  input {
    @apply p-0 pl-2 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }

  input.active {
    outline: none;
    @apply border-primary-600;
  }
  input:active,
  input:focus {
    outline: none;
  }

  input.error:not(:focus) {
    @apply border-destructive text-destructive;
  }

  label {
    @apply font-semibold flex gap-x-1;
  }

  .input-wrapper {
    @apply overflow-hidden;
    @apply flex justify-center items-center;
    @apply bg-surface justify-center;
    @apply border border-gray-300 rounded-[2px];
    @apply cursor-pointer;
    @apply h-8 w-full truncate;
  }

  .input-wrapper:focus-within {
    @apply border-primary-500;
    @apply ring-2 ring-primary-100;
  }
</style>
