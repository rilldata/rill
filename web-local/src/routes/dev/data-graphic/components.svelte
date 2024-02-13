<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { Axis } from "@rilldata/web-common/components/data-graphic/guides";
  import type { AxisSide } from "@rilldata/web-common/components/data-graphic/guides/types";
  const width = 132;
  const height = 132;
  const leftRight = 36;
  const topBottom = 24;
  const sides: AxisSide[] = ["left", "right", "top", "bottom"];
</script>

<h1 class="text-xl">Graphic Components</h1>

<h2 class="text-lg">Axis</h2>
<section>
  {#each sides as axis}
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
