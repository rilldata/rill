<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { AlertDialog as AlertDialogPrimitive } from "bits-ui";
  import { X } from "lucide-svelte";
  import * as AlertDialog from "web-common/src/components/alert-dialog/index.js";

  type $$Props = AlertDialogPrimitive.ContentProps & {
    noCancel?: boolean;
  };

  let className: $$Props["class"] = undefined;
  export let noCancel = false;
  export { className as class };
</script>

<AlertDialog.Portal>
  <AlertDialog.Overlay />
  <AlertDialogPrimitive.Content
    class={cn(
      "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border border-gray-300 bg-surface-subtle p-6 shadow-lg sm:rounded-lg md:w-full",
      className,
    )}
    {...$$restProps}
  >
    <slot />
    {#if !noCancel}
      <AlertDialogPrimitive.Cancel
        class="absolute right-4 top-4 rounded-md opacity-70 transition-opacity hover:opacity-100 focus:outline-none disabled:pointer-events-none"
      >
        <X size={16} class="h-4 w-4" />
        <span class="sr-only">Close</span>
      </AlertDialogPrimitive.Cancel>
    {/if}
  </AlertDialogPrimitive.Content>
</AlertDialog.Portal>
