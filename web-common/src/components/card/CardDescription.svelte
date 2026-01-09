<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";

  export let className: string = "";
</script>

<p
  class={cn(
    "text-slate-500 font-normal text-xs text-center absolute top-[208px]",
    className,
  )}
>
  <slot />
</p>
