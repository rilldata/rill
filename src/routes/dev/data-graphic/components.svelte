<script lang="ts">
  import Axis from "$lib/components/data-graphic/guides/Axis.svelte";
  import SimpleDataGraphic from "$lib/components/data-graphic/SimpleDataGraphic.svelte";
  const width = 132;
  const height = 132;
  const leftRight = 36;
  const topBottom = 24;
</script>

<h1 class="text-xl">Graphic Components</h1>

<h2 class="text-lg">Axis</h2>
<section>
  {#each ["left", "right", "top", "bottom"] as axis}
    <div>
      side=<i>"{axis}"</i>
      <div class="border border-gray-200 rounded">
        <SimpleDataGraphic
          xType="number"
          yType="number"
          {width}
          {height}
          left={leftRight}
          right={leftRight}
          top={topBottom}
          bottom={topBottom}
          xMin={0}
          xMax={100}
          yMin={0}
          yMax={100}
        >
          <Axis side={axis} />
        </SimpleDataGraphic>
      </div>
    </div>
  {/each}
</section>

<style>
  section {
    display: flex;
    flex-direction: row;
    gap: 1rem;
  }
</style>
