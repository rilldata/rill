<script lang="ts">
  import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
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
      <div {...props}>
        {@render children?.()}
      </div>
    {/snippet}
  </DropdownMenuPrimitive.Trigger>
{:else}
  <DropdownMenuPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </DropdownMenuPrimitive.Trigger>
{/if}
