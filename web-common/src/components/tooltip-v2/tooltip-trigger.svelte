<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    ...restProps
  }: TooltipPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <TooltipPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </TooltipPrimitive.Trigger>
{:else}
  <TooltipPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </TooltipPrimitive.Trigger>
{/if}
