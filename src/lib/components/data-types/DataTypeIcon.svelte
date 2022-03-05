<script lang="ts">
import { DATA_TYPE_COLORS, CATEGORICALS } from "$lib/duckdb-data-types";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
export let type;

function typeToSymbol(fieldType:string) {
    return fieldType.slice(0,1);
}
</script>
<Tooltip location="left" distance={16}>
<div
title="{type}"
class="
    { DATA_TYPE_COLORS[type]?.textClass || ''}
    { DATA_TYPE_COLORS[type]?.bgClass || '' }
    grid place-items-center rounded" 
    style="font-size:8px; width: 16px; height: 16px;">
    <div>
        {typeToSymbol(type)}                    
    </div> 
</div>
    <TooltipContent slot="tooltip-content">
        {type}
    </TooltipContent>
</Tooltip>