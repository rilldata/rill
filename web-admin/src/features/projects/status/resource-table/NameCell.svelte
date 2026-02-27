<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let name: string;
</script>

<Tooltip distance={8} location="top">
  <div class="truncate">
    {name}
  </div>
  <TooltipContent slot="tooltip-content">
    <span class="text-xs text-gray-50 font-medium">{name}</span>
  </TooltipContent>
</Tooltip>
