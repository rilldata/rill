<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
  } from "@rilldata/web-common/lib/time/timezone";

  import { IANAZone } from "luxon";

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

  $: watermarkTs = watermark.getTime();
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <button use:builder.action {...builder} class="flex items-center gap-x-1">
      {getAbbreviationForIANA(watermark, activeTimeZone)}
      <span class="flex-none">
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#each availableTimeZones as option (option)}
      {@const zone = new IANAZone(option)}
      <DropdownMenu.CheckboxItem
        class="flex items-center gap-x-1 text-xs cursor-pointer"
        role="menuitem"
        checked={activeTimeZone === option}
        on:click={() => {
          onSelectTimeZone(option);
        }}
      >
        <b class="w-9">
          {getAbbreviationForIANA(watermark, option)}
        </b>
        <p class="inline-block italic w-20">
          GMT {zone.formatOffset(watermarkTs, "short")}
        </p>
        <p>
          {option}
        </p>
      </DropdownMenu.CheckboxItem>
      {#if option === UTCIana}
        <DropdownMenu.Separator />
      {/if}
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
