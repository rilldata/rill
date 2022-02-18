<script lang="ts">
import { DATA_TYPE_ICON_STYLES, CATEGORICALS } from "$lib/duckdb-data-types";
export let type;

function typeToSymbol(fieldType:string) {
    if (CATEGORICALS.has(fieldType)) {
        return "C";
    } else {
        return fieldType.slice(0,1);
    }
}
</script>
<div
title="{type}"
class="
    { DATA_TYPE_ICON_STYLES[type]?.textClass || ''}
    { DATA_TYPE_ICON_STYLES[type]?.bgClass || '' }
    text-ellipsis overflow-hidden whitespace-nowrap 
    grid place-items-center rounded" 
    style="font-size:8px; width: 16px; height: 16px;">
    <div style="transform: translateY(.1px);">
        {typeToSymbol(type)}                    
    </div> 
</div>