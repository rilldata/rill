<script lang="ts">
  import { Checkbox as CheckboxPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Check } from "lucide-svelte";

  type $$Props = CheckboxPrimitive.Props & {
    label?: string;
    inverse?: boolean;
  };

  let className: $$Props["class"] = undefined;
  export let checked: $$Props["checked"] = undefined;
  export let disabled: $$Props["disabled"] = undefined;
  export let label: $$Props["label"] = undefined;
  export let inverse = false;
  export { className as class };
</script>

<div class="flex gap-x-2 {inverse ? 'flex-row-reverse justify-end' : ''}">
  {#if label}
    <label for={$$props.id} class="text-gray-600">{label}</label>
  {/if}
  <CheckboxPrimitive.Root
    {...$$restProps}
    bind:checked
    {disabled}
    class={cn(
      "h-4 w-4 shrink-0 rounded-sm border border-gray-300 bg-transparent ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary-400 data-[state=checked]:border-primary-400",
      className,
    )}
  >
    <CheckboxPrimitive.Indicator
      class={cn("flex items-center justify-center text-white")}
    >
      <Check class="h-3.5 w-3.5" />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
</div>
