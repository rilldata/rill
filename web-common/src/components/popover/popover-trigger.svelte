<script lang="ts">
  import { Popover as PopoverPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    ...restProps
  }: PopoverPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <PopoverPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </PopoverPrimitive.Trigger>
{:else}
  <PopoverPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </PopoverPrimitive.Trigger>
{/if}
