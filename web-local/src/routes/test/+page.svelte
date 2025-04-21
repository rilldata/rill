<script lang="ts">
  // $: ({ canvasEntity } = getCanvasStore(canvasName));
  // $: ({ canvasSpec: _canvasSpec } = canvasEntity.spec);
  // $: ({ instanceId } = $runtime);

  // $: themeFromUrl = $page.url.searchParams.get("theme");
  // $: canvasSpec = $_canvasSpec;

  // $: themeName = themeFromUrl ?? canvasSpec?.theme;
  // $: if (themeName) theme = useTheme(instanceId, themeName);

  import { Bezier } from "bezier-js";

  import chroma from "chroma-js";

  let x = 10;
  let y = 1;
  let minContrast = 1.163;
  let maxContrast = 16.25;
  $: bezier = new Bezier(0, minContrast, x, y, 10, maxContrast);

  $: contrastLut = bezier.getLUT(10);

  $: targetContrastCurve = contrastLut.map((c) => {
    return c.y;
  });

  let sat = {
    light: 0.04,
    primary: 0.4,
    dark: 0.08,
  };

  $: saturationBezier = new Bezier(0, sat.light, 5, sat.primary, 10, sat.dark);

  $: satLut = saturationBezier.getLUT(10);

  $: targetSaturationCurve = satLut.map((c) => {
    return c.y;
  });

  const okay = ["red", "orange", "yellow", "green", "blue", "indigo", "violet"];

  $: blackToShade = chroma
    .scale([chroma("black"), chroma("red")])
    .mode("oklab")
    .gamma(1)
    .colors(22, null);
  // .slice(1, -1);

  $: whiteToShade = chroma
    .scale([chroma("white"), chroma("red")])
    .mode("oklab")
    .gamma(1)
    .colors(22, null);
  // .slice(1, -1);

  import * as Plot from "@observablehq/plot";

  import {
    gammaSpline,
    xs,
    mincontrast,
  } from "@rilldata/web-common/features/themes/actions";

  // Plot.plot({
  //   marks: [Plot.lineY(satLut)],
  // });

  let div: HTMLDivElement;

  // onMount(() => {
  //   const plot = Plot.plot({
  //     marks: [Plot.lineY(satLut)],
  //   });

  //   div?.appendChild(plot);
  // });

  // const points = Array.from({ length: 360 }).map((_, i) => {
  //   return {
  //     x: i,
  //     y: gammaSpline.at(i),
  //   };
  // });

  const points = xs.map((x) => {
    return {
      x: x,
      y: gammaSpline.at(x),
    };
  });

  const minContrastPoints = xs.map((x, i) => {
    return {
      x: x,
      y: mincontrast[i],
    };
  });

  $: {
    div?.firstChild?.remove();
    div?.appendChild(
      Plot.plot({
        y: { domain: [0, 2] },
        marks: [
          Plot.lineY(minContrastPoints, {
            x: "x",
            y: "y",
            marker: true,
            curve: "basis",
          }),
          Plot.lineY(points, { x: "x", y: "y", marker: true, curve: "basis" }),
        ],
      }),
    );
  }
  // let contrastDiv: HTMLDivElement;

  // $: {
  //   contrastDiv?.firstChild?.remove();
  //   contrastDiv?.appendChild(
  //     Plot.plot({
  //       y: { domain: [0, 20] },
  //       marks: [
  //         Plot.lineY(contrastLut, {
  //           x: "x",
  //           y: "y",
  //           domain: [0, 10],
  //           marker: true,
  //         }),
  //         Plot.text(contrastLut, {
  //           x: "x",
  //           y: "y",
  //           text: (d) => d.y.toFixed(3),
  //           // dx: 5,
  //           dy: -10,
  //         }),
  //       ],
  //     }),
  //   );
  // }

  let color = "000000";
  // import chroma from "chroma-js"
  // import * as Plot from "@observablehq/plot";

  const YELLOW = chroma("yellow");
  const VIOLET = chroma("violet");
  const BLACK = chroma("black");
  const WHITE = chroma("white");

  const yellowSpectrum = chroma
    .scale([BLACK, YELLOW, WHITE])
    .mode("oklab")
    .colors(102, null);

  const violetSpectrum = chroma
    .scale([BLACK, VIOLET, WHITE])
    .mode("oklab")
    .colors(102, null);

  $: yellowBackground = yellowSpectrum[15];
  $: violetBacground = violetSpectrum[17];

  $: yellowForeground = yellowSpectrum.find((c) => {
    return chroma.contrast(chroma(c), yellowBackground) > 12;
  });

  $: violetForeground = violetSpectrum.find((c) => {
    return chroma.contrast(chroma(c), violetBacground) > 12;
  });

  // const reds = chroma.deltaE
  let lum = 0.2;
  let hue = 280;
  let saturation = 0.04;

  let red = chroma.oklch(0.63, 0.24, 29.23);
  // import { xs } from "@rilldata/web-common/features/themes/actions";
</script>

<div class="flex font-bold gap-x-2">
  <div class="bg-red-50 size-12 rounded-md grid place-content-center">
    <div class="text-red-700">{chroma("red").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-orange-50 size-12 rounded-md grid place-content-center">
    <div class="text-orange-700">{chroma("orange").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-yellow-50 size-12 rounded-md grid place-content-center">
    <div class="text-yellow-700">{chroma("yellow").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-green-50 size-12 rounded-md grid place-content-center">
    <div class="text-green-700">{chroma("green").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-cyan-50 size-12 rounded-md grid place-content-center">
    <div class="text-cyan-700">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-blue-50 size-12 rounded-md grid place-content-center">
    <div class="text-blue-700">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-indigo-50 size-12 rounded-md grid place-content-center">
    <div class="text-indigo-700">{chroma("indigo").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-violet-50 size-12 rounded-md grid place-content-center">
    <div class="text-violet-700">{chroma("violet").oklch()[2].toFixed(2)}</div>
  </div>
</div>

<div class="flex font-semibold gap-x-2 pt-2">
  <div class="bg-red-500 size-12 rounded-md grid place-content-center">
    <div class="text-red-950">{chroma("red").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-orange-500 size-12 rounded-md grid place-content-center">
    <div class="text-orange-950">{chroma("orange").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-yellow-500 size-12 rounded-md grid place-content-center">
    <div class="text-yellow-950">{chroma("yellow").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-green-500 size-12 rounded-md grid place-content-center">
    <div class="text-green-950">{chroma("green").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-cyan-500 size-12 rounded-md grid place-content-center">
    <div class="text-cyan-950">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-blue-500 size-12 rounded-md grid place-content-center">
    <div class="text-blue-950">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-indigo-500 size-12 rounded-md grid place-content-center">
    <div class="text-indigo-950">{chroma("indigo").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-violet-500 size-12 rounded-md grid place-content-center">
    <div class="text-violet-950">{chroma("violet").oklch()[2].toFixed(2)}</div>
  </div>
</div>

<div class="flex font-semibold gap-x-2 pt-2">
  <div class="bg-red-300 size-12 rounded-md grid place-content-center">
    <div class="text-red-800">{chroma("red").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-orange-300 size-12 rounded-md grid place-content-center">
    <div class="text-orange-800">{chroma("orange").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-yellow-300 size-12 rounded-md grid place-content-center">
    <div class="text-yellow-800">{chroma("yellow").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-green-300 size-12 rounded-md grid place-content-center">
    <div class="text-green-800">{chroma("green").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-cyan-300 size-12 rounded-md grid place-content-center">
    <div class="text-cyan-800">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-blue-300 size-12 rounded-md grid place-content-center">
    <div class="text-blue-800">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-indigo-300 size-12 rounded-md grid place-content-center">
    <div class="text-indigo-800">{chroma("indigo").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-violet-300 size-12 rounded-md grid place-content-center">
    <div class="text-violet-800">{chroma("violet").oklch()[2].toFixed(2)}</div>
  </div>
</div>

<div class="flex font-semibold gap-x-2 pt-2">
  <div class="bg-red-950 size-12 rounded-md grid place-content-center">
    <div class="text-red-300">{chroma("red").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-orange-950 size-12 rounded-md grid place-content-center">
    <div class="text-orange-300">{chroma("orange").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-yellow-950 size-12 rounded-md grid place-content-center">
    <div class="text-yellow-300">{chroma("yellow").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-green-950 size-12 rounded-md grid place-content-center">
    <div class="text-green-300">{chroma("green").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-cyan-950 size-12 rounded-md grid place-content-center">
    <div class="text-cyan-300">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-blue-950 size-12 rounded-md grid place-content-center">
    <div class="text-blue-300">{chroma("blue").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-indigo-950 size-12 rounded-md grid place-content-center">
    <div class="text-indigo-300">{chroma("indigo").oklch()[2].toFixed(2)}</div>
  </div>
  <div class="bg-violet-950 size-12 rounded-md grid place-content-center">
    <div class="text-violet-300">{chroma("violet").oklch()[2].toFixed(2)}</div>
  </div>
</div>

<div>
  <input type="color" bind:value={color} />
  <div
    class="size-12 aspect-square rounded-md border"
    style:background-color={chroma(color).css("oklch")}
  />
  <!-- {chroma(color)?.oklch()} -->
  {chroma.contrastAPCA(chroma(color), chroma("black"))}
</div>

<div bind:this={div}></div>
