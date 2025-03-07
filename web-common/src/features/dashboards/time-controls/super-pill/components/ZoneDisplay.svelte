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
    {#if abbreviation}
      <p class="w-12 truncate overflow-hidden">
        {abbreviation}
      </p>
    {/if}

    <p class="truncate font-normal">
      {#if isBrowserTime}
        Browser Time
      {:else}
        {iana}
      {/if}
    </p>
  </div>

  <span
    class="self-end text-slate-500 rounded-sm line-clamp-1 w-fit flex-none text-[11px]"
  >
    {offset}
  </span>
</div>

<style>
  span {
    font-family: "Source Code Variable";
  }
</style>
