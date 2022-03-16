<script lang="ts">
import { DATA_TYPE_COLORS, 
    CATEGORICALS, 
    NUMERICS, 
    TIMESTAMPS,
    INTEGERS,
    FLOATS, 
    BOOLEANS } from "$lib/duckdb-data-types";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

import CategoricalType from "../icons/CategoricalType.svelte";
import TimestampType from "../icons/TimestampType.svelte";
import BooleanType from "../icons/BooleanType.svelte";
import IntegerType from "../icons/IntegerType.svelte";
import FloatType from "../icons/FloatType.svelte";
export let color = 'text-gray-400';
export let type;
export let suppressTooltip = false;

function typeToSymbol(fieldType:string) {
    if (INTEGERS.has(fieldType)) {
        return IntegerType;
    } else if (FLOATS.has(fieldType)) {
        return FloatType;
    } else if (CATEGORICALS.has(fieldType)) {
        return CategoricalType;
    } else if (TIMESTAMPS.has(fieldType)) {
        return TimestampType;
    } else if (BOOLEANS.has(fieldType)) {
        return BooleanType;
    }
}
</script>
<Tooltip location="left" distance={16} suppress={suppressTooltip}>
<div
title="{type}"
class="
    {color}
    grid place-items-center rounded" 
    style="width: 16px; height: 16px;">
    <div>
        <svelte:component this={typeToSymbol(type)} size="16px" />
    </div> 
</div>
    <TooltipContent slot="tooltip-content">
        {type}
    </TooltipContent>
</Tooltip>