<script lang="ts">
  import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
  export let tableName;
  export let importName;
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(50,0,0,.6), rgba(50,0,30,.8))"
>
  <div slot="title">
    <span class="font-bold">Importing table</span>
    <span class="italic">{importName}</span>
    <div>as <span class="font-bold">{tableName}</span></div>
  </div>
</BlockingOverlayContainer>
