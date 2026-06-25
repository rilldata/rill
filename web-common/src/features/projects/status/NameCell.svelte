<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let name: string;
</script>

<Tooltip distance={8}>
  <div class="truncate">
    {name}
  </div>
  <TooltipContent slot="tooltip-content">
    {name}
  </TooltipContent>
</Tooltip>
