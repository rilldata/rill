<script lang="ts">
  import { Tabs as TabsPrimitive } from "bits-ui";
  import { cn } from "../../lib/shadcn";

  type $$Props = TabsPrimitive.TriggerProps;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export { className as class };
</script>

<TabsPrimitive.Trigger
  class={cn(
    "inline-flex items-center justify-center whitespace-nowrap text-sm font-medium text-muted-foreground transition-all",
    "pb-2 rounded-none border-b-2 border-transparent",
    "data-[state=active]:border-primary-500 data-[state=active]:text-foreground",
    "ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
    "disabled:pointer-events-none disabled:opacity-50",
    className,
  )}
  {value}
  {...$$restProps}
  on:click
  on:keydown
  on:focus
>
  <slot />
</TabsPrimitive.Trigger>
