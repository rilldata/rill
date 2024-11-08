<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
  } from "@rilldata/web-common/lib/time/timezone";
  import ZoneDisplay from "./ZoneDisplay.svelte";

  // watermark indicates the latest reference point in the dashboard
  export let watermark: Date;
  export let availableTimeZones: string[];
  export let activeTimeZone: string;
  export let onSelectTimeZone: (timeZone: string) => void;

  const userLocalIANA = getLocalIANA();
  const UTCIana = "UTC";

  let open = false;

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
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      class="flex items-center gap-x-1"
      aria-label="Timezone selector"
    >
      {getAbbreviationForIANA(watermark, activeTimeZone)}
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#each availableTimeZones as option (option)}
      <DropdownMenu.CheckboxItem
        checked={activeTimeZone === option}
        on:click={() => {
          onSelectTimeZone(option);
        }}
      >
        <ZoneDisplay iana={option} {watermark} />
      </DropdownMenu.CheckboxItem>
      {#if option === UTCIana}
        <DropdownMenu.Separator />
      {/if}
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
