<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  import { Select as SelectPrimitive } from "bits-ui";
  import { Check } from "lucide-svelte";

  type $$Props = SelectPrimitive.ItemProps & {
    description?: string;
  };
  // type $$Events = Required<SelectPrimitive.ItemEvents>;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export let label: $$Props["label"] = undefined;
  export let description: $$Props["description"] = undefined;
  export let disabled: $$Props["disabled"] = undefined;
  export { className as class };
</script>

<SelectPrimitive.Item
  {value}
  {disabled}
  {label}
  class={cn(
    "group relative flex flex-col w-full cursor-pointer select-none items-center text-fg-primary rounded-sm py-1.5 px-2 text-sm outline-none data-[highlighted]:bg-popover-accent data-[highlighted]:text-fg-accent data-[disabled]:opacity-50",
    className,
  )}
  {...$$restProps}
>
  <div class="flex flex-row items-center justify-between w-full gap-x-2">
    <slot>
      {label ? label : value}
    </slot>

    <span class="ml-auto flex h-3.5 w-3.5 justify-end">
      <Check class="size-3.5 hidden group-data-[state=checked]:block" />
    </span>
  </div>
  {#if description}
    <div class="text-fg-secondary">
      {description}
    </div>
  {/if}
</SelectPrimitive.Item>
