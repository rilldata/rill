<script lang="ts">
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    ...restProps
  }: DropdownMenuPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <DropdownMenuPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </DropdownMenuPrimitive.Trigger>
{:else}
  <DropdownMenuPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </DropdownMenuPrimitive.Trigger>
{/if}
