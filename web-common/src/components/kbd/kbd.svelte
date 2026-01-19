<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";

  type $$Props = HTMLAttributes<HTMLElement>;

  let className: $$Props["class"] = undefined;
</script>

<kbd
  data-slot="kbd"
  class={cn(
    "bg-muted text-muted-foreground pointer-events-none inline-flex h-5 w-fit min-w-5 items-center justify-center gap-1 rounded-sm px-1 font-sans text-xs font-medium select-none",
    "[&_svg:not([class*='size-'])]:size-3",
    "[[data-slot=tooltip-content]_&]:bg-background/20 [[data-slot=tooltip-content]_&]:text-background dark:[[data-slot=tooltip-content]_&]:bg-background/10",
    className,
  )}
  {...$$restProps}
>
  <slot />
</kbd>
