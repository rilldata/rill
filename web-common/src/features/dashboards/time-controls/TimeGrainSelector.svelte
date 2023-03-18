<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import {
    getTimeGrainFromRuntimeGrain,
    isMinGrainBigger,
  } from "./utils/time-grain";
  import type { TimeGrainOption } from "./utils/time-types";

  export let metricViewName: string;
  export let timeGrainOptions: TimeGrainOption[];
  export let minTimeGrain: V1TimeGrain;

  const dispatch = createEventDispatcher();
  const EVENT_NAME = "select-time-grain";

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeGrain = $dashboardStore?.selectedTimeRange?.interval;

  $: activeTimeGrainPretty =
    getTimeGrainFromRuntimeGrain(activeTimeGrain)?.label;

  $: timeGrains = timeGrainOptions
    ? timeGrainOptions.map((timeGrain) => {
        const isGrainPossible = !isMinGrainBigger(minTimeGrain, timeGrain);
        return {
          main: timeGrain.label,
          disabled: !timeGrain.enabled || !isGrainPossible,
          key: timeGrain.grain,
          description: !timeGrain.enabled
            ? "not valid for this time range"
            : !isGrainPossible
            ? "bigger than min time grain"
            : undefined,
        };
      })
    : undefined;

  const onTimeGrainSelect = (timeGrain: V1TimeGrain) => {
    dispatch(EVENT_NAME, { timeGrain });
  };
</script>

{#if activeTimeGrain && timeGrainOptions}
  <WithSelectMenu
    distance={8}
    options={timeGrains}
    selection={{
      main: activeTimeGrainPretty,
      key: activeTimeGrain,
    }}
    on:select={(event) => onTimeGrainSelect(event.detail.key)}
    let:toggleMenu
    let:active
  >
    <button
      class="px-3 py-2 rounded flex flex-row gap-x-2 hover:bg-gray-200 hover:dark:bg-gray-600"
      on:click={toggleMenu}
    >
      <div>
        Metric trends by <span class="font-bold">{activeTimeGrainPretty}</span>
      </div>
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="16px" />
        </div>
      </IconSpaceFixer>
    </button>
  </WithSelectMenu>
{/if}
