<script lang="ts">
  import { Checkbox as CheckboxPrimitive } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import { Check } from "lucide-svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  type $$Props = CheckboxPrimitive.Props & {
    label?: string;
    inverse?: boolean;
    hint?: string;
    optional?: boolean;
  };

  export let checked: $$Props["checked"] = undefined;
  export let disabled: $$Props["disabled"] = undefined;
  export let label: $$Props["label"] = undefined;
  export let inverse = false;
  export let hint: string | undefined = undefined;
  export let optional: boolean = false;
  export { className as class };

  let className: $$Props["class"] = undefined;
</script>

<div
  class="flex items-center gap-x-1 {inverse
    ? 'flex-row-reverse justify-end'
    : ''}"
>
  <CheckboxPrimitive.Root
    {...$$restProps}
    bind:checked
    {disabled}
    class={cn(
      "h-4 w-4 shrink-0 rounded border border-gray-300 bg-transparent ring-offset-2 ring-offset-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-400 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ",
      // FIXME: bg-base-primary is not in the system, but used in figma
      `data-[state=checked]:bg-[#6366F1] data-[state=checked]:border-transparent`,
      className,
    )}
  >
    <CheckboxPrimitive.Indicator
      class={cn("flex items-center justify-center text-white")}
    >
      <Check class="h-3.5 w-3.5" />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>

  {#if label}
    <label for={$$props.id} class="flex items-center text-sm gap-x-1">
      {label}
      {#if optional}
        <span class="text-gray-500 text-[12px] font-normal capitalize"
          >(optional)</span
        >
      {/if}
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="240px" slot="tooltip-content">
            {@html hint}
          </TooltipContent>
        </Tooltip>
      {/if}
    </label>
  {/if}
</div>
