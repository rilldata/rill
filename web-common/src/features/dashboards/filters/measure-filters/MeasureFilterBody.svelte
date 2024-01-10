<!-- @component
  renders the body content of a filter set chip:
  - a label for the current dimension
  - a certain number of "show" values (default 1)
  - an indication of how many other dimensions are selected past the show values
-->
<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let dimensionName: string;
  export let label: string | undefined;
  export let expr: V1Expression | undefined;
  export let labelMaxWidth = "160px";
  export let active = false;

  $: shortLabel = MeasureFilterOptions.find(
    (o) => o.value === expr?.cond?.op,
  )?.shortLabel;
</script>

<div class="flex gap-x-2">
  <div
    class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
    style:max-width={labelMaxWidth}
  >
    {label} for {dimensionName}
  </div>
  <div class="flex flex-wrap flex-row items-baseline gap-y-1">
    {#if shortLabel}
      {shortLabel} {expr?.cond?.exprs?.[1].val ?? ""}
    {/if}
    <IconSpaceFixer className="pl-1" pullRight>
      <div class="transition-transform" class:-rotate-180={active}>
        <CaretDownIcon className="inline" size="10px" />
      </div>
    </IconSpaceFixer>
  </div>
</div>
