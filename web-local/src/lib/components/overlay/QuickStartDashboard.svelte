<script lang="ts">
  import BlockingOverlayContainer from "@rilldata/web-common/features/temp/BlockingOverlayContainer.svelte";

  export let sourceName: string;
  export let timeDimension: string;
</script>

<BlockingOverlayContainer
  bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
>
  <div slot="title">
    <span class="font-bold">Creating a metrics dashboard for source</span>
    <span>{sourceName}</span>
    <div>
      using <span class="font-bold">{timeDimension}</span> as time dimension column.
    </div>
  </div>
</BlockingOverlayContainer>
