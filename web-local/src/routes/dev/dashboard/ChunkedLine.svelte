<script lang="ts">
  import { contexts } from "@rilldata/web-local/lib/components/data-graphic/constants";
  import { getContext } from "svelte";
  export let data;
  export let xAccessor: string;
  export let yAccessor: string;
  // get the scale functions from the data graphic
  const xScale = getContext(contexts.scale("x"));
  const yScale = getContext(contexts.scale("y"));
</script>

<g>
  <path />
</g>
