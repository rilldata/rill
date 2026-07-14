<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Select as SelectPrimitive } from "bits-ui";

  type $$Props = SelectPrimitive.ContentProps & {
    sameWidth?: boolean;
  };

  let className: $$Props["class"] = undefined;
  export let sideOffset: $$Props["sideOffset"] = 4;
  /** When true, the dropdown width matches the trigger width via CSS variable. */
  export let sameWidth: boolean = false;
  export { className as class };
</script>

<SelectPrimitive.Portal>
  <SelectPrimitive.Content
    {sideOffset}
    class={cn(
      "relative z-50 min-w-[8rem] overflow-hidden rounded-[2px] border bg-popover text-popover-foreground shadow-md focus:outline-none",
      sameWidth && "w-[var(--bits-select-anchor-width)]",
      className,
    )}
    {...$$restProps}
  >
    <div class="w-full p-1">
      <slot />
    </div>
  </SelectPrimitive.Content>
</SelectPrimitive.Portal>
