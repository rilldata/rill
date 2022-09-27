<script lang="ts">
  import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
  export let tableName;
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(0, 0, 0, .5), rgba(0, 0, 0,.6))"
>
  <div slot="title">
    <span class="font-bold">Exporting dataset</span>
    <div class="italic">{tableName}</div>
  </div>
</BlockingOverlayContainer>
