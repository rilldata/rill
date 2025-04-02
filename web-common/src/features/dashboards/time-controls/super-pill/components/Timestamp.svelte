<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { DateTime, Duration } from "luxon";

  export let date: DateTime = DateTime.now();
  export let zone: string;
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
        .filter(([, value]) => value !== 0)
        .slice(0, 2),
    ),
  ).toHuman({
    listStyle: "narrow",
    maximumFractionDigits: 0,
  });
</script>

<Tooltip {suppress}>
  <button
    class:italic
    class="text-xs text-inherit"
    on:click={() => {
      if (isoString) copyToClipboard(isoString);
    }}
  >
    {#if showDate}
      {formattedString}
    {:else}
      ({humanReadableTimeOffset} ago)
    {/if}
  </button>

  <TooltipContent slot="tooltip-content">
    <div class="flex flex-col gap-y-1 items-center">
      <span>
        {#if showDate}
          {#if humanReadableTimeOffset.length}
            {humanReadableTimeOffset} ago
          {/if}
        {:else}
          {formattedString}
        {/if}
      </span>

      <span>
        {isoString}
      </span>
    </div>
  </TooltipContent>
</Tooltip>
