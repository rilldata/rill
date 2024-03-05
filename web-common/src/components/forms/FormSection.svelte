<script lang="ts">
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let title: string;
  export let description: string = "";
  export let tooltip: boolean = false;
</script>

<div class="flex flex-col bg-white p-3 gap-y-3 rounded">
  <div class="flex flex-col">
    <span
      class="flex flex-row items-center gap-1 text-base text-medium text-slate-900"
    >
      <div>{title}</div>
      {#if tooltip}
        <Tooltip>
          <InfoCircle />
          <TooltipContent slot="tooltip-content" maxWidth="600px">
            <slot name="tooltip-content" />
          </TooltipContent>
        </Tooltip>
      {/if}
    </span>
    {#if description}
      <span class="text-sm text-slate-600">{description}</span>
    {/if}
  </div>
  <slot />
</div>
