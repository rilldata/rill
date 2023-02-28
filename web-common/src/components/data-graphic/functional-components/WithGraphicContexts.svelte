<script lang="ts">
  import { getGraphicContexts } from "./get-graphic-contexts";
  const { xScale, yScale, config } = getGraphicContexts();
</script>

<slot xScale={$xScale} yScale={$yScale} config={$config} />
