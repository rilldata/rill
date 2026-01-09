<script lang="ts">
  import { Checkbox as CheckboxPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Check } from "lucide-svelte";

  export let checked: boolean;
  export let label: string;
  export let onCheckedChange: (checked: boolean) => void;

  function toggle() {
    onCheckedChange(!checked);
  }
</script>

<button
  type="button"
  class={cn(
    "flex items-center gap-2 p-3 border rounded-sm cursor-pointer transition-colors duration-150 w-[calc(50%-4px)]",
    checked
      ? "border-primary-400 bg-primary-50"
      : "border-gray-200 hover:border-gray-300 hover:bg-gray-50",
  )}
  on:click={toggle}
>
  <CheckboxPrimitive.Root
    {checked}
    class={cn(
      "h-4 w-4 shrink-0 rounded border border-gray-300 bg-transparent pointer-events-none",
      "data-[state=checked]:bg-primary-500 data-[state=checked]:border-transparent",
    )}
    tabindex={-1}
  >
    <CheckboxPrimitive.Indicator
      class="flex items-center justify-center text-white"
    >
      <Check class="h-3.5 w-3.5" />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>

  <span class="text-sm font-medium text-gray-900">{label}</span>
</button>
