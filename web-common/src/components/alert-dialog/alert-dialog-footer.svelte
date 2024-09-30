<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  type $$Props = HTMLAttributes<HTMLDivElement>;

  let className: $$Props["class"] = undefined;
  export { className as class };
</script>

<div
  class={cn(
    "flex flex-col-reverse gap-y-2 sm:flex-row sm:justify-end sm:space-x-2",
    className,
  )}
  {...$$restProps}
>
  <slot />
</div>
