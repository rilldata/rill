<script lang="ts">
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import { buttonVariants } from "./button-variants";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import type { Snippet } from "svelte";

  let {
    asChild = false,
    children,
    class: className,
    ...restProps
  }: AlertDialogPrimitive.CancelProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <AlertDialogPrimitive.Cancel {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </AlertDialogPrimitive.Cancel>
{:else}
  <AlertDialogPrimitive.Cancel
    class={cn(buttonVariants({ variant: "outline" }), "mt-2 sm:mt-0", className)}
    {...restProps}
  >
    {@render children?.()}
  </AlertDialogPrimitive.Cancel>
{/if}
