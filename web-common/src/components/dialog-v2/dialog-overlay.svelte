<script lang="ts">
  import { Dialog as DialogPrimitive } from "bits-ui";
  import { fade } from "svelte/transition";
  import { cn } from "@rilldata/web-common/lib/shadcn";

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
    "fixed inset-0 bg-gray-400 transition-opacity opacity-40 ",
    className,
  )}
  {...$$restProps}
/>
