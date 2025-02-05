<script lang="ts">
  import { cn, flyAndScale } from "@rilldata/web-common/lib/shadcn";
  import { Dialog as DialogPrimitive } from "bits-ui";
  import Cross2 from "svelte-radix/Cross2.svelte";
  import * as Dialog from "web-common/src/components/dialog-v2/index.js";

  type $$Props = DialogPrimitive.ContentProps & { noClose?: boolean };

  let className: $$Props["class"] = undefined;
  export let transition: $$Props["transition"] = flyAndScale;
  export let transitionConfig: $$Props["transitionConfig"] = {
    duration: 200,
  };
  export let noClose = false;
  export { className as class };
</script>

<Dialog.Portal>
  <Dialog.Overlay />
  <DialogPrimitive.Content
    {transition}
    {transitionConfig}
    class={cn(
      "fixed left-[50%] top-[50%] z-50 grid w-full max-w-xl translate-x-[-50%] translate-y-[-50%] gap-4 border bg-surface p-6 shadow-lg sm:rounded-lg md:w-full",
      className,
    )}
    {...$$restProps}
  >
    <slot />
    {#if !noClose}
      <DialogPrimitive.Close
        class="absolute right-4 top-4 rounded-md opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
      >
        <Cross2 class="h-4 w-4" />
        <span class="sr-only">Close</span>
      </DialogPrimitive.Close>
    {/if}
  </DialogPrimitive.Content>
</Dialog.Portal>
