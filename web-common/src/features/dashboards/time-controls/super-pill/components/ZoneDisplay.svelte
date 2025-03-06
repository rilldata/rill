<script lang="ts">
  import { formatIANA } from "@rilldata/web-common/lib/time/timezone";
  import { DateTime } from "luxon";

  export let iana: string;
  export let isBrowserTime = false;
  export let selected = false;
  export let abbreviation: string | undefined = undefined;
  export let offset: string = formatIANA(iana, DateTime.now()).offset;
</script>

<div
  class="flex items-baseline text-xs w-full cursor-pointer justify-between overflow-hidden"
>
  <div class="flex gap-x-1 w-full overflow-hidden items-baseline">
    <p class="truncate font-normal" class:!font-bold={selected}>
      {#if isBrowserTime}
        Browser Time
      {:else}
        {iana}
      {/if}
    </p>

    {#if abbreviation}
      <p class="min-w-12 max-w-12 text-[10px]">
        {abbreviation}
      </p>
    {/if}
  </div>

  <span
    class="self-end italic line-clamp-1 w-fit flex-none text-xs"
    class:!font-semibold={selected}
  >
    {offset}
  </span>
</div>
