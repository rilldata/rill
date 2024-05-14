<script lang="ts">
  import { line, curveLinear } from "d3-shape";
  import { interpolatePath } from "d3-interpolate-path";
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";

  export let data: { x: number; y: number }[];
  export let type: "secondary" | "primary" | "comparison";
  export let color: string | undefined;

  const lineFunction = line<{ x: number; y: number }>()
    .defined((d) => d.y !== null && d.y !== undefined)
    .x(({ x }) => x)
    .y(({ y }) => y)
    .curve(curveLinear);

  const tweenedPath = tweened(lineFunction(data), {
    duration: 400,
    interpolate: interpolatePath,
    easing: cubicOut,
  });

  $: tweenedPath.set(lineFunction(data)).catch((e) => console.error(e));
</script>

<path
  vector-effect="non-scaling-stroke"
  d={$tweenedPath}
  fill="none"
  class="stroke-{color ?? 'primary-500'}"
  stroke-width={type === "comparison" ? 1.5 : 1}
/>
