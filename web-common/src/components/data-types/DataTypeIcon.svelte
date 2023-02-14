<script lang="ts">
  import BooleanType from "@rilldata/web-common/components/icons/BooleanType.svelte";
  import FloatType from "@rilldata/web-common/components/icons/FloatType.svelte";
  import IntegerType from "@rilldata/web-common/components/icons/IntegerType.svelte";
  import StringlikeType from "@rilldata/web-common/components/icons/StringlikeType.svelte";
  import TimestampType from "@rilldata/web-common/components/icons/TimestampType.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    BOOLEANS,
    FLOATS,
    INTEGERS,
    INTERVALS,
    isList,
    isNested,
    STRING_LIKES,
    TIMESTAMPS,
  } from "@rilldata/web-common/lib/duckdb-data-types";
  import ListType from "../icons/ListType.svelte";
  import StructType from "../icons/StructType.svelte";

  export let color = "text-gray-400";
  export let type;
  export let suppressTooltip = false;

  function typeToSymbol(fieldType: string) {
    if (INTEGERS.has(fieldType)) {
      return IntegerType;
    } else if (FLOATS.has(fieldType)) {
      return FloatType;
    } else if (STRING_LIKES.has(fieldType)) {
      return StringlikeType;
    } else if (TIMESTAMPS.has(fieldType) || INTERVALS.has(type)) {
      return TimestampType;
    } else if (BOOLEANS.has(fieldType)) {
      return BooleanType;
    } else if (isList(fieldType)) {
      return ListType;
    } else if (isNested(fieldType)) {
      return StructType;
    }
  }
</script>

<Tooltip location="left" distance={16} suppress={suppressTooltip}>
  <div
    title={type}
    class="
    {color}
    grid place-items-center rounded"
    style="width: 16px; height: 16px;"
  >
    <div>
      <svelte:component this={typeToSymbol(type)} size="16px" />
    </div>
  </div>
  <TooltipContent slot="tooltip-content">
    {type}
  </TooltipContent>
</Tooltip>
