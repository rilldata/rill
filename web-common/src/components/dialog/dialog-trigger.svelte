<script lang="ts">
  import { Dialog as DialogPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  // svelte-ignore custom_element_props_identifier
  let {
    asChild = false,
    children,
    ...restProps
  }: DialogPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <DialogPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </DialogPrimitive.Trigger>
{:else}
  <DialogPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </DialogPrimitive.Trigger>
{/if}
