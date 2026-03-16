<script lang="ts">
  import { Tooltip as TooltipPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
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
      <div {...props}>
        {@render children?.()}
      </div>
    {/snippet}
  </TooltipPrimitive.Trigger>
{:else}
  <TooltipPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </TooltipPrimitive.Trigger>
{/if}
