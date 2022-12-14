<script lang="ts">
  import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
  export let tableName;
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(0, 0, 0, .5), rgba(0, 0, 0,.6))"
>
  <div slot="title">
    <span class="ui-copy-strong">Exporting dataset</span>
    <div>{tableName}</div>
  </div>
</BlockingOverlayContainer>
