<script lang="ts">
  import { Switch as SwitchPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = SwitchPrimitive.Props & {
    small?: boolean;
  };
  // type $$Events = SwitchPrimitive.Events;

  let className: $$Props["class"] = undefined;
  export let checked: $$Props["checked"] = undefined;
  export { className as class };
  export let small = false;
</script>

<SwitchPrimitive.Root
  {...$$restProps}
  bind:checked
  class={cn(
    "peer inline-flex h-[20px] w-[36px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary-400 data-[state=unchecked]:bg-input",
    small && "h-3 w-[22px]",
    className,
  )}
  on:click
  on:keydown
>
  <SwitchPrimitive.Thumb
    class={cn(
      "pointer-events-none block h-4 w-4 rounded-full bg-background shadow-lg ring-0 transition-transform data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0",
      small &&
        "h-2.5 w-2.5 data-[state=checked]:translate-x-[9px] data-[state=unchecked]:-translate-x-[1px]",
    )}
  />
</SwitchPrimitive.Root>
