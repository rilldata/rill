<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  export let source: string;
</script>

<Tooltip location="right" alignment="start" distance={8}>
  <NavigationEntry>
    <svelte:fragment slot="name">
      <div
        class="w-full text-ellipsis overflow-hidden whitespace-nowrap "
        style="direction: rtl;
      text-align: left;"
      >
        {source}<span class="text-gray-400">/..</span>
      </div>
    </svelte:fragment>
  </NavigationEntry>
  <TooltipContent slot="tooltip-content" maxWidth="300px">
    <TooltipTitle>
      <svelte:fragment slot="name">{source}</svelte:fragment>
    </TooltipTitle>
    {source}
  </TooltipContent>
</Tooltip>
