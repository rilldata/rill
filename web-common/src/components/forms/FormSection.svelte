<script lang="ts">
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let title: string;
  export let description: string = "";
  export let padding = "p-3";
  export let showSectionToggle = false;
  export let enabled = true;
</script>

<div class="flex flex-col bg-white {padding} gap-y-3 rounded">
  <div class="flex flex-col">
    <span
      class="flex flex-row items-center gap-1 text-base text-medium text-slate-900"
    >
      <div>{title}</div>
      {#if $$slots["tooltip-content"]}
        <Tooltip>
          <InfoCircle />
          <TooltipContent slot="tooltip-content" maxWidth="600px">
            <slot name="tooltip-content" />
          </TooltipContent>
        </Tooltip>
      {/if}
      {#if showSectionToggle}
        <div class="grow"></div>
        <Switch bind:checked={enabled} />
      {/if}
    </span>
    {#if $$slots["description"]}
      <slot name="description" />
    {:else if description}
      <span class="text-sm text-slate-600">{description}</span>
    {/if}
  </div>
  {#if enabled}
    <slot />
  {/if}
</div>
