<script lang="ts">
  import BlockingOverlayContainer from "../../../layout/BlockingOverlayContainer.svelte";
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
>
  <div slot="title">
    <span class="font-bold">Preparing source for import</span>
  </div>
</BlockingOverlayContainer>
