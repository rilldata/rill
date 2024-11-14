<script lang="ts">
  import { cn } from "@rilldata/web-common/lib/shadcn";

  import { Select as SelectPrimitive } from "bits-ui";
  import { Check } from "lucide-svelte";

  type $$Props = SelectPrimitive.ItemProps;
  // type $$Events = Required<SelectPrimitive.ItemEvents>;

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"];
  export let label: $$Props["label"] = undefined;
  export let disabled: $$Props["disabled"] = undefined;
  export { className as class };
</script>

<SelectPrimitive.Item
  {value}
  {disabled}
  {label}
  class={cn(
    "relative flex w-full cursor-pointer select-none items-center rounded-sm py-1.5 px-2 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:opacity-50",
    className,
  )}
  {...$$restProps}
  on:click
  on:pointermove
  on:focusin
>
  <slot>
    {label ? label : value}
  </slot>

  <span class="ml-auto flex flex-none h-3.5 w-3.5 items-center justify-center">
    <SelectPrimitive.ItemIndicator>
      <Check class="size-3.5 " />
    </SelectPrimitive.ItemIndicator>
  </span>
</SelectPrimitive.Item>
