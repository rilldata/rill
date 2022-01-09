<script>
import IconButton from "$lib/components/IconButton.svelte";
import RowIcon from "$lib/components/icons/RowIcon.svelte";
import JSONIcon from "$lib/components/icons/JsonIcon.svelte";

// Icon Button Row.
let iconRowSelected = 0;

</script>

<div class="grid grid-flow-col justify-start bg-white p-4">
    <IconButton selected={iconRowSelected === 0} on:click={() => { iconRowSelected = 0;}}> <RowIcon /> </IconButton>
    <IconButton selected={iconRowSelected === 1} on:click={() => { iconRowSelected = 1;}}> <JSONIcon /> </IconButton>
</div>