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
  }: AlertDialogPrimitive.ActionProps & {
    asChild?: boolean;
    children?: Snippet;
  } = $props();
</script>

{#if asChild}
  <AlertDialogPrimitive.Action {...restProps}>
    {#snippet child({ props })}
      <span style="display:contents" {...props}>
        {@render children?.()}
      </span>
    {/snippet}
  </AlertDialogPrimitive.Action>
{:else}
  <AlertDialogPrimitive.Action
    class={cn(buttonVariants(), className)}
    {...restProps}
  >
    {@render children?.()}
  </AlertDialogPrimitive.Action>
{/if}
