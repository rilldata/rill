<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import { isGrainBigger, TIME_GRAIN } from "./utils/time-grain";
  import type { TimeGrainOption } from "./utils/time-types";

  export let metricViewName: string;
  export let timeGrainOptions: TimeGrainOption[];
  export let minTimeGrain: V1TimeGrain;

  const dispatch = createEventDispatcher();
  const EVENT_NAME = "select-time-grain";

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeGrain = $dashboardStore?.selectedTimeRange?.interval;

  $: activeTimeGrainLabel = TIME_GRAIN[activeTimeGrain].label;

  $: timeGrains = timeGrainOptions
    ? timeGrainOptions
        .filter(
          (timeGrain) =>
            !isGrainBigger(minTimeGrain, timeGrain.grain) && timeGrain.enabled
        )
        .map((timeGrain) => {
          return {
            main: timeGrain.label,
            key: timeGrain.grain,
          };
        })
    : undefined;

  const onTimeGrainSelect = (timeGrain: V1TimeGrain) => {
    dispatch(EVENT_NAME, { timeGrain });
  };
</script>

{#if activeTimeGrain && timeGrainOptions}
  <WithSelectMenu
    paddingTop={1}
    paddingBottom={1}
    minWidth="160px"
    distance={8}
    options={timeGrains}
    selection={{
      main: activeTimeGrainLabel,
      key: activeTimeGrain,
    }}
    on:select={(event) => onTimeGrainSelect(event.detail.key)}
    let:toggleMenu
    let:active
  >
    <button
      class:bg-gray-200={active}
      class="px-3 py-2 rounded flex flex-row gap-x-2 items-center hover:bg-gray-200 hover:dark:bg-gray-600"
      on:click={toggleMenu}
    >
      <div>
        Metric trends by <span class="font-bold">{activeTimeGrainLabel}</span>
      </div>
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="14px" />
        </div>
      </IconSpaceFixer>
    </button>
  </WithSelectMenu>
{/if}
