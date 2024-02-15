<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import SelectorButton from "@rilldata/web-common/features/dashboards/time-controls/SelectorButton.svelte";
  import {
    getAbbreviationForIANA,
    getLabelForIANA,
    getLocalIANA,
    getTimeZoneNameFromIANA,
  } from "@rilldata/web-common/lib/time/timezone";
  import { createEventDispatcher } from "svelte";
  import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";

  export let metricViewName: string;
  // now indicates the latest reference point in the dashboard
  export let now: Date;
  export let availableTimeZones: string[];

  const dispatch = createEventDispatcher();
  const userLocalIANA = getLocalIANA();
  const UTCIana = "UTC";

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  // Filter out user time zone and UTC
  $: availableTimeZones = availableTimeZones.filter(
    (tz) => tz !== userLocalIANA && tz !== UTCIana,
  );

  // Add local and utc time zone to the top of the list
  $: availableTimeZones = [userLocalIANA, UTCIana, ...availableTimeZones];

  // If localIANA is same as UTC, remove UTC from the list
  $: if (userLocalIANA === UTCIana) {
    availableTimeZones = availableTimeZones.slice(1);
  }

  const onTimeZoneSelect = (timeZone: string) => {
    dispatch("select-time-zone", { timeZone });
  };
</script>

{#if activeTimeZone}
  <WithTogglableFloatingElement
    alignment="start"
    distance={8}
    let:toggleFloatingElement
    let:active
  >
    <Tooltip distance={8} suppress={active}>
      <SelectorButton
        {active}
        label="Timezone selector"
        on:click={() => {
          toggleFloatingElement();
        }}
      >
        <div class="flex items-center gap-x-2">
          <Globe size="16px" />
          <span class="font-bold"
            >{getAbbreviationForIANA(now, activeTimeZone)}</span
          >
        </div>
      </SelectorButton>
      <TooltipContent slot="tooltip-content" maxWidth="220px">
        Select a time zone for the dashboard.
        <br />
        Currently using {getTimeZoneNameFromIANA(now, activeTimeZone)}.
      </TooltipContent>
    </Tooltip>
    <Menu
      slot="floating-element"
      let:toggleFloatingElement
      on:click-outside={toggleFloatingElement}
      on:escape={toggleFloatingElement}
      minWidth="320px"
    >
      {#each availableTimeZones as option}
        {@const label = getLabelForIANA(now, option)}
        <MenuItem
          icon
          selected={activeTimeZone === option}
          on:select={() => {
            onTimeZoneSelect(option);
            toggleFloatingElement();
          }}
        >
          <svelte:fragment slot="icon">
            {#if option === activeTimeZone}
              <Check size="20px" color="#15141A" />
            {:else}
              <Spacer size="20px" />
            {/if}
          </svelte:fragment>
          <span>
            <span class="inline-block font-bold w-9">
              {label.abbreviation}
            </span>
            <span class="inline-block italic w-20">
              {label.offset}
            </span>
            {label.iana}
          </span>
        </MenuItem>
        {#if option === UTCIana}
          <Divider />
        {/if}
      {/each}
    </Menu>
  </WithTogglableFloatingElement>
{/if}
