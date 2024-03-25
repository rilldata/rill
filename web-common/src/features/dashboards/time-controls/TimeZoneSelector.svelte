<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
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

  let open = false;

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

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <SelectorButton
        builders={[builder]}
        active={open}
        label="Timezone selector"
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
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#each availableTimeZones as option}
      {@const label = getLabelForIANA(now, option)}
      <DropdownMenu.CheckboxItem
        class="flex items-center gap-x-1 text-xs cursor-pointer"
        role="menuitem"
        checked={activeTimeZone === option}
        on:click={() => {
          onTimeZoneSelect(option);
        }}
      >
        <b class="w-9">
          {label.abbreviation}
        </b>
        <p class="inline-block italic w-20">
          {label.offset}
        </p>
        <p>{label.iana}</p>
      </DropdownMenu.CheckboxItem>
      {#if option === UTCIana}
        <DropdownMenu.Separator />
      {/if}
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
