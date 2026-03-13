<script lang="ts">
  import { ContextMenu as ContextMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    ...restProps
  }: ContextMenuPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <ContextMenuPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <div {...props}>
        {@render children?.()}
      </div>
    {/snippet}
  </ContextMenuPrimitive.Trigger>
{:else}
  <ContextMenuPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </ContextMenuPrimitive.Trigger>
{/if}
