<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = SelectPrimitive.TriggerProps;
  // type $$Events = SelectPrimitive.TriggerEvents;

  let className: $$Props["class"] = undefined;

  export let el: HTMLButtonElement | undefined = undefined;
  export { className as class };
</script>

<SelectPrimitive.Trigger
  bind:el
  class={cn(
    "flex h-8 w-full items-center justify-between rounded-[2px] border border-gray-300 bg-transparent px-2 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:border-primary-400 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1",
    className,
  )}
  {...$$restProps}
>
  <slot />
  <div class="caret transition-transform">
    <CaretDownIcon size="12px" className="fill-gray-600" />
  </div>
</SelectPrimitive.Trigger>

<style lang="postcss">
  :global(button[aria-expanded="true"] > .caret) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
