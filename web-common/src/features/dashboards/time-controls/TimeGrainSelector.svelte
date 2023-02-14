<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { prettyTimeGrain, TimeGrainOption } from "./time-range-utils";

  export let selectedTimeGrain: V1TimeGrain;
  export let availableTimeGrains: V1TimeGrain[];
  export let selectableTimeGrains: TimeGrainOption[];

  const dispatch = createEventDispatcher();
  const EVENT_NAME = "select-time-grain";

  $: options = selectableTimeGrains
    ? selectableTimeGrains.map(({ timeGrain, enabled }) => {
        const isTimeGrainAvailable =
          !availableTimeGrains.length ||
          availableTimeGrains.includes(timeGrain);
        return {
          main: prettyTimeGrain(timeGrain),
          disabled: !enabled || !isTimeGrainAvailable,
          key: timeGrain,
          description: !enabled
            ? "not valid for this time range"
            : !isTimeGrainAvailable
            ? "not available"
            : undefined,
        };
      })
    : undefined;

  const onTimeGrainSelect = (timeGrain: V1TimeGrain) => {
    dispatch(EVENT_NAME, { timeGrain });
  };
</script>

{#if selectedTimeGrain && selectableTimeGrains}
  <WithSelectMenu
    {options}
    selection={{
      main: prettyTimeGrain(selectedTimeGrain),
      key: selectedTimeGrain,
    }}
    on:select={(event) => onTimeGrainSelect(event.detail.key)}
    let:toggleMenu
    let:active
  >
    <button
      class="px-4 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600 transition-tranform duration-100"
      on:click={toggleMenu}
    >
      <span class="font-bold"
        >by {prettyTimeGrain(selectedTimeGrain)} increments</span
      >
      <span class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon size="16px" />
      </span>
    </button>
  </WithSelectMenu>
{/if}
