<script lang="ts">
import { DATA_TYPE_COLORS, CATEGORICALS, NUMERICS, TIMESTAMPS, BOOLEANS } from "$lib/duckdb-data-types";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

import CategoricalType from "../icons/CategoricalType.svelte";
import NumericType from "../icons/NumericType.svelte";
import TimestampType from "../icons/TimestampType.svelte";
import BooleanType from "../icons/BooleanType.svelte";

export let type;

function typeToSymbol(fieldType:string) {
    //return fieldType.slice(0,1);
    if (CATEGORICALS.has(fieldType)) {
        return CategoricalType;
    } else if (NUMERICS.has(fieldType)) {
        return NumericType;
    } else if (TIMESTAMPS.has(fieldType)) {
        return TimestampType;
    } else if (BOOLEANS.has(fieldType)) {
        return BooleanType;
    }
}
</script>
<Tooltip location="left" distance={16}>
<div
title="{type}"
class="
    text-gray-400
    grid place-items-center rounded" 
    style="width: 16px; height: 16px;">
    <div>
        <svelte:component this={typeToSymbol(type)} size="16px" />
        <!-- {typeToSymbol(type)}                     -->
    </div> 
</div>
    <TooltipContent slot="tooltip-content">
        {type}
    </TooltipContent>
</Tooltip>