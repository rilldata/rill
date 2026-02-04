<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Dialog as DialogPrimitive } from "bits-ui";
  import { fade } from "svelte/transition";

  type $$Props = DialogPrimitive.OverlayProps;

  let className: $$Props["class"] = undefined;
  export let transition: $$Props["transition"] = fade;
  export let transitionConfig: $$Props["transitionConfig"] = {
    duration: 150,
  };
  export { className as class };
</script>

<DialogPrimitive.Overlay
  {transition}
  {transitionConfig}
  class={cn(
    "fixed inset-0 bg-neutral-800 dark:bg-neutral-100 transition-opacity opacity-60 z-50",
    className,
  )}
  {...$$restProps}
/>
