<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";
  import { createEventDispatcher } from "svelte";
  import { useDashboardStore } from "../dashboard-stores";
  import { getLabelForIANA } from "@rilldata/web-common/lib/time/timezone";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";

  export let metricViewName: string;
  export let now: Date;

  export let timeZoneOptions = [
    {
      label: getLabelForIANA(now, "Etc/UTC"),
      zoneIANA: "Etc/UTC",
    },
    { label: getLabelForIANA(now, "Asia/Kolkata"), zoneIANA: "Asia/Kolkata" },
    {
      label: getLabelForIANA(now, "America/Los_Angeles"),
      zoneIANA: "America/Los_Angeles",
    },
  ];

  const dispatch = createEventDispatcher();

  $: dashboardStore = useDashboardStore(metricViewName);
  $: activeTimeZone = $dashboardStore?.selectedTimezone;

  $: timeZoneOptions = timeZoneOptions.map((option) => ({
    ...option,
    key: option.zoneIANA,
    main: option.label,
  }));

  const onTimeZoneSelect = (timeZone: string) => {
    dispatch("select-time-zone", { timeZone });
  };
</script>

{#if activeTimeZone && timeZoneOptions}
  <WithSelectMenu
    paddingTop={1}
    paddingBottom={1}
    minWidth="160px"
    distance={8}
    options={timeZoneOptions}
    selection={{
      main: getLabelForIANA(now, activeTimeZone),
      key: activeTimeZone,
    }}
    on:select={(event) => onTimeZoneSelect(event.detail.key)}
    let:toggleMenu
    let:active
  >
    <button
      class:bg-gray-200={active}
      class="px-3 py-2 rounded flex flex-row gap-x-2 items-center hover:bg-gray-200 hover:dark:bg-gray-600"
      on:click={toggleMenu}
    >
      <div class="flex items-center gap-x-2">
        <Globe />
        <span class="font-bold">{getLabelForIANA(now, activeTimeZone)}</span>
      </div>
      <IconSpaceFixer pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon size="14px" />
        </div>
      </IconSpaceFixer>
    </button>
  </WithSelectMenu>
{/if}
