<script lang="ts">
import HistogramBase from "./HistogramBase.svelte";

interface HistogramBin {
    bucket:number;
    low:number;
    high:number;
    count:number;
}

export let data:HistogramBin[];
export let width = 60;
export let height = 19;
export let fillColor = 'hsl(340, 70%, 70%)';
export let baselineStrokeColor:string;

// scales

</script>

<HistogramBase 
    {data} {width} {height} top={4} bottom={4} left={0} right={0} buffer={0} {fillColor} {baselineStrokeColor} separate={false}
/>
