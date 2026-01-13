<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Command as CommandPrimitive } from "cmdk-sv";
  import Search from "../icons/Search.svelte";

  // type $$Props = CommandPrimitive.InputProps;

  let className: string | undefined | null = undefined;
  export { className as class };
  export let value: string = "";
</script>

<div
  class="flex items-center border border-gray-200 bg-gray-50 has-[:focus]:border-primary-400 rounded-md px-3 disabled:cursor-not-allowed disabled:opacity-50"
  data-cmdk-input-wrapper=""
>
  <Search size="16" className="mr-2 shrink-0" />
  <CommandPrimitive.Input
    class={cn(
      "bg-transparent flex w-full py-1 text-xs outline-none",
      className,
    )}
    {...$$restProps}
    bind:value
  />
</div>
