<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";
  import { prettyTimeGrain, TimeGrainOption } from "./time-range-utils";

  export let metricViewName: string;
  export let timeGrainOptions: TimeGrainOption[];

  const dispatch = createEventDispatcher();
  const EVENT_NAME = "select-time-grain";

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeGrain = $dashboardStore?.selectedTimeRange?.interval;

  $: timeGrains = timeGrainOptions
    ? timeGrainOptions.map(({ timeGrain, enabled }) => ({
        main: prettyTimeGrain(timeGrain),
        disabled: !enabled,
        key: timeGrain,
        description: !enabled ? "not valid for this time range" : undefined,
      }))
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
      main: prettyTimeGrain(activeTimeGrain),
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
      <span class="font-bold"
        >by {prettyTimeGrain(activeTimeGrain)} increments</span
      >
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="16px" />
        </div>
      </IconSpaceFixer>
    </button>
  </WithSelectMenu>
{/if}
