<script lang="ts">
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";

  export let sorted: boolean;
  export let selected: boolean;
  export let excludeMode: boolean;
  export let isBeingCompared: boolean;
  export let value: string | number | boolean | undefined;
</script>

<div
  class:sorted
  class="flex flex-row gap-x-1 size-full text-right items-center"
>
  <span class="h-full aspect-square bg-white items-center flex">
    {#if selected}
      {#if !excludeMode && isBeingCompared}
        <CheckCircle className={"bg-red-400"} size="18px" />
      {:else if !excludeMode}
        <Check size="20px" />
      {:else if excludeMode}
        <Cancel size="20px" />
      {/if}
    {/if}
  </span>
  <p class="truncate size-full text-left label items-center flex">
    {value}
  </p>
</div>
