<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
</script>

<!-- FIXME: Revisit when we can add new user group  -->
<div class="flex flex-row items-center gap-x-1">
  <Tooltip location="bottom" alignment="middle" distance={8}>
    <div class="text-yellow-600">
      <InfoCircle size="16px" />
    </div>
    <TooltipContent maxWidth="400px" slot="tooltip-content">
      This person can still edit because they are part of the group
    </TooltipContent>
  </Tooltip>
  <slot />
</div>
