<!-- @component
  renders the body content of a filter set chip:
  - a label for the current measure
  - a short hand notation of the filter criteria
-->
<script lang="ts">
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { MeasureFilterOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { V1Operation } from "@rilldata/web-common/runtime-client";

  export let dimensionName: string;
  export let label: string | undefined;
  export let expr: V1Expression | undefined;
  export let labelMaxWidth = "160px";
  export let active = false;
  export let readOnly = false;

  let shortLabel: string | undefined;
  $: if (expr?.cond?.op) {
    if (
      expr?.cond?.op === V1Operation.OPERATION_AND ||
      expr?.cond?.op === V1Operation.OPERATION_OR
    ) {
      shortLabel = `${
        expr?.cond?.op === V1Operation.OPERATION_OR ? "!" : ""
      }(${JSON.stringify(
        expr?.cond?.exprs?.[0]?.cond?.exprs?.[1]?.val ?? "",
      )}, ${JSON.stringify(
        expr?.cond?.exprs?.[1]?.cond?.exprs?.[1]?.val ?? "",
      )})`;
    } else {
      shortLabel =
        MeasureFilterOptions.find((o) => o.value === expr?.cond?.op)
          ?.shortLabel +
        " " +
        JSON.stringify(expr?.cond?.exprs?.[1].val);
    }
  }
</script>

<div class="flex gap-x-2">
  <div
    class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
    style:max-width={labelMaxWidth}
  >
    {label}
    {#if dimensionName}
      <!-- span needed to make sure the space before the `for` is not removed by prettier -->
      <span> for {dimensionName}</span>
    {/if}
  </div>
  <div class="flex flex-wrap flex-row items-baseline gap-y-1">
    {#if shortLabel}
      {shortLabel}
    {/if}
    {#if !readOnly}
      <IconSpaceFixer className="pl-2" pullRight>
        <div class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon className="inline" size="10px" />
        </div>
      </IconSpaceFixer>
    {/if}
  </div>
</div>
