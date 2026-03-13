<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    ...restProps
  }: AlertDialogPrimitive.TriggerProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <AlertDialogPrimitive.Trigger {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </AlertDialogPrimitive.Trigger>
{:else}
  <AlertDialogPrimitive.Trigger {...restProps}>
    {@render children?.()}
  </AlertDialogPrimitive.Trigger>
{/if}
