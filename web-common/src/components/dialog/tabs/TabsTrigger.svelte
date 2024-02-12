<script lang="ts">
  import { Tabs as TabsPrimitive } from "bits-ui";
  import { cn } from "../../../lib/shadcn";
  import NumberedCircle from "./NumberedCircle.svelte";

  type $$Props = TabsPrimitive.TriggerProps & { tabIndex: number };
  // type $$Events = TabsPrimitive.TriggerEvents;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export let tabIndex: number; // Ideally, we'd instead use `builder["tabindex"]`, but it's 0 for the active tab and -1 for the inactive tabs
  export { className as class };
</script>

<TabsPrimitive.Trigger
  class={cn(
    "flex gap-x-2 p-2 items-center justify-center whitespace-nowrap w-[200px] h-[34px] text-sm text-gray-500 font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none data-[state=active]:bg-background data-[state=active]:text-foreground",
    className,
  )}
  {value}
  {...$$restProps}
  disabled
  let:builder
>
  <NumberedCircle
    number={tabIndex + 1}
    bgColor={builder["data-state"] === "active" ? "bg-gray-800" : "bg-gray-400"}
  />
  <slot />
</TabsPrimitive.Trigger>
