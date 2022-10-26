<script lang="ts">
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let description = "";
  export let hint = "";
  export let href = "";
</script>

<div class="flex flex-row items-center pl-1 gap-x-1">
  <div class="text-gray-500 italic">
    {description}
  </div>
  {#if hint}
    <Tooltip location="bottom" alignment="middle" distance={8}>
      <button
        class="text-gray-500"
        style="transform:translateY(-.5px)"
        on:click|preventDefault={() => {
          window.open(href, "_blank");
        }}
      >
        <InfoCircle size="13px" />
      </button>
      <TooltipContent maxWidth="400px" slot="tooltip-content">
        {@html hint}
      </TooltipContent>
    </Tooltip>
  {/if}
</div>
