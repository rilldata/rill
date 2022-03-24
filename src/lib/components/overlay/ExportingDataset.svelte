<script lang="ts">
import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
export let tableName;
export let path;
</script>

<BlockingOverlayContainer bg="linear-gradient(to right, rgba(0,0,50, .5), rgba(0, 30, 30,.6))">

    <div slot="title">
        <span class='font-bold'>Exporting dataset</span>
        <div class="italic">{tableName}</div> 
        <div>to <span class='font-bold'>{path}</span></div>
    </div>
</BlockingOverlayContainer>
