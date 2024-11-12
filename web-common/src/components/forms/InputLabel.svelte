<script lang="ts">
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";

  export let id: string;
  export let label: string;
  export let hint: string | undefined = undefined;
  export let optional = false;
</script>

<div class="label-wrapper">
  <label for={id} class="line-clamp-1 capitalize">
    {label}
    {#if optional}
      <span class="text-gray-500 text-[12px] font-normal">(optional)</span>
    {/if}
  </label>
  {#if hint}
    <Tooltip location="right" alignment="middle" distance={8}>
      <div class="text-gray-500">
        <InfoCircle size="13px" />
      </div>
      <TooltipContent maxWidth="240px" slot="tooltip-content">
        {@html hint}
      </TooltipContent>
    </Tooltip>
  {/if}
  <slot name="mode-switch" />
</div>

<style lang="postcss">
  .label-wrapper {
    @apply flex items-center gap-x-1;
  }

  label {
    @apply text-sm font-medium text-gray-800;
  }
</style>
