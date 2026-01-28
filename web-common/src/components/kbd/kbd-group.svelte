<script lang="ts">
  import type { HTMLAttributes } from "svelte/elements";
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";

  type $$Props = HTMLAttributes<HTMLElement>;

  let className: $$Props["class"] = undefined;
</script>

<kbd
  data-slot="kbd-group"
  class={cn("inline-flex items-center gap-1", className)}
  {...$$restProps}
>
  <slot />
</kbd>
