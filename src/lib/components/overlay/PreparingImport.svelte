<script lang="ts">
  import BlockingOverlayContainer from "./BlockingOverlayContainer.svelte";
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(50,0,0,.6), rgba(50,0,30,.8))"
>
  <div slot="title">
    <span class="font-bold">Preparing source for import</span>
  </div>
</BlockingOverlayContainer>
