<script lang="ts">
  import DOMPurify from "dompurify";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id: string;
  export let label: string;
  export let hint: string | undefined = undefined;
  export let optional = false;
  export let capitalize = true;
  export let faint = false;
  export let link: string | undefined = undefined;
  export let small = false;
</script>

<div class="label-wrapper">
  <label
    for={id}
    class="line-clamp-1 {small ? 'text-xs' : 'text-sm'} text-fg-secondary"
    class:capitalize
    class:faint
  >
    {label}
    {#if optional}
      <span class="text-fg-tertiary text-[12px] font-normal">(optional)</span>
    {/if}
  </label>
  {#if hint}
    <Tooltip location="right" alignment="middle" distance={8}>
      <svelte:element
        this={link ? "a" : "div"}
        {...link ? { href: link, target: "_blank" } : {}}
        class="text-icon-muted"
      >
        <InfoCircle size="13px" />
      </svelte:element>
      <TooltipContent maxWidth="240px" slot="tooltip-content">
        {@html DOMPurify.sanitize(hint)}
      </TooltipContent>
    </Tooltip>
  {/if}
  <slot name="mode-switch" />
</div>

<style lang="postcss">
  .label-wrapper {
    @apply flex items-center gap-x-1 select-none;
  }

  label {
    @apply font-medium;
  }

  label.faint {
    @apply text-fg-secondary;
  }
</style>
