<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import Link from "../icons/Link.svelte";
  // import { Link2, Link2Off } from "lucide-svelte";
  import LinkBreak from "svelte-radix/LinkBreak2.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  // import Link2 from "svelte-radix/Link1.svelte";

  type $$Props = SelectPrimitive.TriggerProps & {
    lockable?: boolean;
    lockTooltip?: string;
    // See: https://www.bits-ui.com/docs/components/select#selecttrigger
    // Converts div to button tag
    class?: string;
  };
  // type $$Events = SelectPrimitive.TriggerEvents;

  let className: $$Props["class"] = undefined;

  export let el: HTMLButtonElement | undefined = undefined;
  export let lockable = false;
  export let lockTooltip = "";
  export { className as class };

  let locked = lockable;
</script>

<SelectPrimitive.Trigger
  bind:el
  disabled={locked}
  class={cn(
    "flex h-8 w-full items-center relative justify-between rounded-[2px] border border-gray-300 bg-transparent px-2 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:border-primary-400 disabled:cursor-not-allowed disabled:bg-gray-50 [&>span]:line-clamp-1",
    className,
  )}
  {...$$restProps}
>
  <slot />
  {#if locked}
    <Tooltip>
      <button
        on:click={() => {
          locked = false;
        }}
        class="group active:bg-gray-50 grid bg-background place-content-center h-full absolute right-0 w-[40px] border-l pointer-events-auto cursor-pointer"
      >
        <Link size="14px" className="fill-gray-600 group-hover:hidden" />
        <LinkBreak
          class="text-primary-600 hidden group-hover:block"
          size="14px"
        />
      </button>

      <TooltipContent slot="tooltip-content">
        {lockTooltip}
      </TooltipContent>
    </Tooltip>
  {/if}
  <div class="caret transition-transform">
    <CaretDownIcon size="12px" className="fill-gray-600" />
  </div>
</SelectPrimitive.Trigger>

<style lang="postcss">
  :global(button[aria-expanded="true"] > .caret) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
