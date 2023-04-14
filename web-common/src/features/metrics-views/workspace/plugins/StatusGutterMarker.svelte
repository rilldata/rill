<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  export let level: string;
</script>

<Tooltip distance={8}>
  <div
    style:height="16px"
    style:width="24px"
    class:bg-red-200={level === "error"}
    class="grid justify-center items-center"
  >
    <!-- svelte-ignore missing-declaration -->
    <Cancel />
  </div>
  <TooltipContent slot="tooltip-content">yes!!!</TooltipContent>
</Tooltip>
