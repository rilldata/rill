<script>
export let type;

function typeToSymbol(fieldType) {
    if (fieldType === 'BYTE_ARRAY' || fieldType === 'VARCHAR') {
        return {symbol: "C", text: 'categorical (BYTE_ARRAY)', color: 'red'};
    } else {
        return {symbol: fieldType.slice(0,1), text: fieldType, color: 'sky'};
    }
}
</script>
<div
class:bg-sky-200={type === 'VARCHAR'}
class:bg-red-200={!(type === 'VARCHAR')}
class:bg-teal-300={type === 'TIMESTAMP'}
class:text-sky-800={type === 'VARCHAR'}
class:text-red-800={type === 'VARCHAR'}
class:text-teal-800={type === 'TIMESTAMP'}
class="
    text-ellipsis overflow-hidden whitespace-nowrap 
    grid place-items-center rounded" 
    style="font-size:8px; width: 16px; height: 16px;">
    <div style="transform: translateY(.2px);">
        {typeToSymbol(type).symbol}                    
    </div> 
</div>