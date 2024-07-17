<script lang="ts">
  import { cn, flyAndScale } from "@rilldata/web-common/lib/shadcn";
  import { Popover as PopoverPrimitive } from "bits-ui";

  type $$Props = PopoverPrimitive.ContentProps;

  let className: $$Props["class"] = undefined;
  export let transition: $$Props["transition"] = flyAndScale;
  export let transitionConfig: $$Props["transitionConfig"] = undefined;
  export let align: $$Props["align"] = "center";
  export let sideOffset: $$Props["sideOffset"] = 4;
  export { className as class };
</script>

<PopoverPrimitive.Content
  {transition}
  {transitionConfig}
  {align}
  {sideOffset}
  {...$$restProps}
  class={cn(
    "z-50 w-72 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none",
    className,
  )}
>
  <slot />
</PopoverPrimitive.Content>
