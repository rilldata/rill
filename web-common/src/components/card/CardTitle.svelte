<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";

  export let position: "middle" | "bottom" = "bottom";
  export let className: string = "";
</script>

<h3
  class={cn(
    "text-slate-600 font-medium text-sm",
    position === "bottom" ? "absolute top-[180px]" : "",
    className,
  )}
>
  <slot />
</h3>
