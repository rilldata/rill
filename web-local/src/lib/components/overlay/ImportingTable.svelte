<script lang="ts">
  import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
  export let tableName;
  export let importName;
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
>
  <div slot="title">
    <span class="ui-copy-strong">Importing source</span>
    <span>{importName}</span>
    <div>as <span class="ui-copy-strong">{tableName}</span></div>
  </div>
</BlockingOverlayContainer>
