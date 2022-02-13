<script>
import Varchar from "./Varchar.svelte";
import Number from "./Number.svelte";
import Timestamp from "./Timestamp.svelte";

export let type = 'VARCHAR';
export let isNull = false;
export let inTable = false;

let dataType = Varchar;
$: {
    if (type === 'VARCHAR') {
        dataType = Varchar;
    } else if (type === 'TIMESTAMP') {
        dataType = Timestamp;
    } else {
        dataType = Number;
    }
}
</script>

<svelte:component this={dataType} {isNull} {inTable}>
    <slot />
</svelte:component>