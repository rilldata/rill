<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";

  export let showTooltip = false;
</script>

<div class="flex flex-row items-center gap-x-1">
  {#if showTooltip}
    <Tooltip location="bottom" alignment="middle" distance={8}>
      <div class="text-yellow-600">
        <InfoCircle size="16px" />
      </div>
      <TooltipContent maxWidth="400px" slot="tooltip-content">
        This person can still edit because they are part of the group
        INSERT_GROUP_NAME
      </TooltipContent>
    </Tooltip>
  {/if}
  <slot />
</div>
