<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import NavigationEntry from "@rilldata/web-local/lib/components/navigation/NavigationEntry.svelte";
  export let uri: string;
  export let abbreviatedURI = uri;
</script>

<NavigationEntry>
  <svelte:fragment slot="name">
    <div
      class="w-full text-ellipsis overflow-hidden whitespace-nowrap "
      style="direction: rtl;
      text-align: left;"
    >
      {abbreviatedURI}<span class="text-gray-400">/..</span>
    </div>
  </svelte:fragment>
  <TooltipContent slot="tooltip-content" maxWidth="300px">
    <TooltipTitle>
      <svelte:fragment slot="name">{uri}</svelte:fragment>
    </TooltipTitle>
  </TooltipContent>
</NavigationEntry>
