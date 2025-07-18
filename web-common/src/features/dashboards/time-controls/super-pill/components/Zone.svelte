<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import type { DateTime } from "luxon";
  import ZoneContent from "./ZoneContent.svelte";

  // watermark indicates the latest reference point in the dashboard
  export let watermark: DateTime;
  export let availableTimeZones: string[];
  export let activeTimeZone: string;
  export let context: string;
  export let lockTimeZone = false;
  export let onSelectTimeZone: (timeZone: string) => void;

  let open = false;
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      class="flex items-center gap-x-1"
      aria-label="Timezone selector"
      title={!availableTimeZones.length ? "No timezones configured" : ""}
      disabled={lockTimeZone}
    >
      {getAbbreviationForIANA(watermark, activeTimeZone)}
      {#if !lockTimeZone}
        <span class="flex-none transition-transform" class:-rotate-180={open}>
          <CaretDownIcon />
        </span>
      {/if}
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-80">
    <ZoneContent
      {watermark}
      {availableTimeZones}
      {activeTimeZone}
      {context}
      {onSelectTimeZone}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
