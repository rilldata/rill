<script lang="ts">
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { DateTime, Duration } from "luxon";
  import { Tooltip } from "bits-ui";

  export let date: DateTime = DateTime.now();
  export let zone: string;
  export let id: string;
  export let showDate = true;
  export let suppress = false;
  export let italic = false;

  $: zonedDate = date.setZone(zone);
  $: isoString = zonedDate.toISO();

  $: formattedString = zonedDate.toLocaleString(
    DateTime.DATETIME_MED_WITH_WEEKDAY,
  );

  $: humanReadableTimeOffset = Duration.fromObject(
    Object.fromEntries(
      Object.entries(DateTime.now().diff(date).rescale().toObject())
        .filter(([label, value]) => value !== 0 && label !== "milliseconds")
        .slice(0, 2),
    ),
  ).toHuman({
    listStyle: "narrow",
    maximumFractionDigits: 0,
  });
</script>

<Tooltip.Root disableHoverableContent={true}>
  <Tooltip.Trigger asChild let:builder id="{id}-timestamp-trigger">
    <button
      use:builder.action
      {...builder}
      class:italic
      class="text-xs text-inherit"
      on:click={() => {
        if (isoString) copyToClipboard(isoString);
      }}
    >
      {#if showDate}
        {formattedString}
      {:else}
        {humanReadableTimeOffset} ago
      {/if}
    </button>
  </Tooltip.Trigger>
  <Tooltip.Content
    hidden={suppress}
    class="w-fit flex"
    side="right"
    sideOffset={16}
  >
    <div
      class="flex flex-col gap-y-1 items-center flex-none bg-gray-700 dark:bg-gray-900 shadow-md text-surface rounded p-2 pt-1 pb-1"
    >
      {#if showDate}
        {#if humanReadableTimeOffset.length}
          <span>{humanReadableTimeOffset} ago</span>
        {/if}
      {:else}
        <span>{formattedString}</span>
      {/if}

      <span>
        {isoString}
      </span>
    </div>
  </Tooltip.Content>
</Tooltip.Root>
