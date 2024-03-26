<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = SelectPrimitive.TriggerProps;
  // type $$Events = SelectPrimitive.TriggerEvents;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<SelectPrimitive.Trigger
  class={cn(
    "flex h-8 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-transparent px-2 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary-400 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1",
    className,
  )}
  {...$$restProps}
>
  <slot />
  <div>
    <CaretDownIcon size="12px" className="fill-gray-600" />
  </div>
</SelectPrimitive.Trigger>
