<script>
import { NUMERICS, TIMESTAMPS } from "$lib/duckdb-data-types"
import Varchar from "./Varchar.svelte";
import Number from "./Number.svelte";
import Timestamp from "./Timestamp.svelte";

export let type = 'VARCHAR';
export let isNull = false;
export let inTable = false;
export let dark = false;

let dataType = Varchar;
$: {
    if (NUMERICS.has(type)) {
        dataType = Number;
    } else if (TIMESTAMPS.has(type)) {
        dataType = Timestamp;
    } else {
        // default to the varchar style
        dataType = Varchar;
    }
}
</script>

<svelte:component this={dataType} {isNull} {inTable} {dark}>
    <slot />
</svelte:component>