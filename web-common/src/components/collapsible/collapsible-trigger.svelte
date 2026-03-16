<script lang="ts">
  import { Collapsible as CollapsiblePrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    asChild = false,
    children,
    ...restProps
  }: CollapsiblePrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <CollapsiblePrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </CollapsiblePrimitive.Trigger>
{:else}
  <CollapsiblePrimitive.Trigger {...restProps}>
    {@render children?.()}
  </CollapsiblePrimitive.Trigger>
{/if}
