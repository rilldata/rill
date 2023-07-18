<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { useDashboardStore } from "../dashboard-stores";
  import {
    getAbbreviationForIANA,
    getLabelForIANA,
    getLocalIANA,
  } from "@rilldata/web-common/lib/time/timezone";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import SelectorButton from "@rilldata/web-common/features/dashboards/time-controls/SelectorButton.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-common/components/menu";
  import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";

  export let metricViewName: string;
  // now indicates the latest reference point in the dashboard
  export let now: Date;
  export let timeZoneOptions = DEFAULT_TIMEZONES;

  const dispatch = createEventDispatcher();
  const userLocalIANA = getLocalIANA();
  const UTCIana = "Etc/UTC";

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  // Filter out user time zone and UTC
  $: timeZoneOptions = timeZoneOptions.filter(
    (tz) => tz !== userLocalIANA && tz !== UTCIana
  );

  // Add local and utc time zone to the top of the list
  $: timeZoneOptions = [userLocalIANA, UTCIana, ...timeZoneOptions];

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
        on:click={() => {
          toggleFloatingElement();
        }}
      >
        <div class="flex items-center gap-x-2">
          <Globe />
          <span class="font-bold"
            >{getAbbreviationForIANA(now, activeTimeZone)}</span
          >
        </div>
      </SelectorButton>
      <TooltipContent slot="tooltip-content" maxWidth="220px">
        Select a reference time zone for the dashboard
      </TooltipContent>
    </Tooltip>
    <Menu
      slot="floating-element"
      on:escape={toggleFloatingElement}
      label="Timezone selector"
      minWidth="320px"
    >
      {#each timeZoneOptions as option}
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
