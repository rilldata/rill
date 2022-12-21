<!--
@component
Takes in a domain and range, and outputs in a slot the derived linear scale
-->
<script lang="ts">
  import { scaleLinear } from "d3-scale";
  export let domain: Date[] | number[];
  export let range: number[];
  $: scale = scaleLinear().domain(domain).range(range);
</script>

<slot {scale} />
