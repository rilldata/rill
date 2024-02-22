<script lang="ts">
  import { Tabs as TabsPrimitive } from "bits-ui";
  import { cn } from "../../../lib/shadcn";

  type $$Props = TabsPrimitive.ContentProps & {
    tabIndex: number;
    currentTabIndex: number;
  };

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export let tabIndex: $$Props["tabIndex"];
  // The tab index that is currently selected in the parent Tabs component.
  export let currentTabIndex: $$Props["currentTabIndex"];
  export { className as class };
</script>

<TabsPrimitive.Content
  class={cn(
    "ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
    className,
  )}
  {value}
  {...$$restProps}
>
  <!-- Prevent the slotted content from mounting when the user is viewing a different tab. -->
  {#if tabIndex === currentTabIndex}
    <slot />
  {/if}
</TabsPrimitive.Content>
