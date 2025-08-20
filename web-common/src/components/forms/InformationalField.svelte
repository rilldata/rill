<script lang="ts">
  import InfoCircle from "../icons/InfoCircle.svelte";
  import LongDescription from "../tooltip/LongDescription.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let description = "";
  export let hint = "";
  export let href = "";
</script>

<div class="flex flex-row items-center gap-x-1">
  <div class="text-gray-500">
    {description}
  </div>
  {#if hint}
    <Tooltip location="bottom" alignment="middle" distance={8}>
      <a
        class="text-gray-500 hover:cursor-pointer"
        style="transform:translateY(-.5px)"
        {href}
        target="_blank"
        rel="noreferrer noopener"
      >
        <InfoCircle size="13px" />
      </a>
      <TooltipContent slot="tooltip-content">
        <LongDescription>
          {@html hint}
        </LongDescription>
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
