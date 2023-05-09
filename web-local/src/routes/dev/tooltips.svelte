<script>
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import SummaryAndHistogram from "@rilldata/web-local/lib/components/viz/histogram/NumericHistogram.svelte";
  import { data01 } from "./_hist-data.ts";
  let x = 50;
  let y = 500;
  let location = "left";
  let alignment = "middle";
  let hoverElement;

  let activated = true;

  function changeActivation(event) {
    if (event.key === "Escape") {
      activated = !activated;
    }
  }

  function handleDrag(event) {
    if (activated) {
      x = Math.max(
        20,
        Math.min(event.clientX - hoverElement.clientWidth / 2 + scrollX)
      );
      y = Math.max(
        20,
        Math.min(event.clientY - hoverElement.clientHeight / 2 + scrollY)
      );
    }
    // }
  }

  // show programmatic tooltip generation

  let showTooltip = true;
  // setInterval(() => {
  //     showTooltip = !showTooltip;
  // }, 1000)

  $: if (
    (location === "top" || location === "bottom") &&
    (alignment === "top" || alignment === "bottom")
  ) {
    alignment = "center";
  }

  $: if (
    (location === "left" || location === "right") &&
    (alignment === "left" || alignment === "right")
  ) {
    alignment = "center";
  }
</script>

<svelte:window on:keydown={changeActivation} />
<svelte:body on:mousemove={handleDrag} />

<div
  class="fixed p-3 bg-white border border-black border-3 rounded"
  style:right="20px"
  style:top="20px"
>
  {activated ? "activated!" : "not activated ..."}
  <div>Hit ESC to toggle.</div>
</div>

<div
  style:height="200vh"
  style:width="200vw"
  class="border border-8 border-black overflow-hidden"
>
  <h1 class="text-xl p-5 fixed left-2 top-2 bg-white backdrop-blur-md">
    Tooltip Positioning Semantics.
  </h1>
  <div class="mt-24" />
  <Tooltip alignment="left">
    <div style:width="max-content">position</div>

    <div slot="tooltip-content" class="bg-black text-white p-5">
      another element
    </div>
  </Tooltip>

  <Tooltip alignment="left" bind:active={showTooltip} distance={16}>
    <button class="bg-red-500 text-white p-3 rounded">position</button>
    <TooltipContent slot="tooltip-content">another element</TooltipContent>
  </Tooltip>

  <p>
    Contrary to popular belief, Lorem Ipsum is not simply random text. It has
    roots in a piece of classical Latin literature from 45 BC, making it over
    2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney
    College in Virginia, looked up one of the more obscure Latin words,
    consectetur, from a Lorem Ipsum passage, and going through the cites of the
    word in classical literature, discovered the undoubtable source. Lorem Ipsum
    comes from sections 1.10.32 and 1.10.33 of "de Finibus Bonorum et Malorum"
    (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a
    treatise on the theory of
    <Tooltip>
      <span class="font-bold"> ethics, </span>
      <div slot="tooltip-content" class="bg-black text-white p-5">
        another element!
      </div>
    </Tooltip>
    very popular during the Renaissance. The first line of Lorem Ipsum, "Lorem ipsum
    dolor sit amet..", comes from a line in section 1.10.32.
  </p>

  <div class="m-12">
    <select bind:value={location}>
      <option value="left">left</option>
      <option value="right">right</option>
      <option value="top">top</option>
      <option value="bottom">bottom</option>
    </select>

    <select bind:value={alignment}>
      <option value="start">start</option>
      <option value="middle">middle</option>
      <option value="end">end</option>
    </select>
  </div>

  <Tooltip distance={20} {location} {alignment}>
    <div
      on:mousemove={handleDrag}
      class="absolute rounded outline p-5 hover:outline-sky-500 hover:text-sky-800 hover:bg-sky-100 transition-colors hover:outline-4 hover:font-bold backdrop-blur-sm select-none"
      style:left="{x}px"
      style:top="{y}px"
      style:width="200px"
      bind:this={hoverElement}
    >
      drag this one around 🚐
    </div>

    <!-- <TooltipContent> -->
    <div
      slot="tooltip-content"
      style="min-height: 300px; width: 300px;"
      class="border border-2 border-black p-3 rounded backdrop-blur-md"
    >
      This is a tooltip. It will follow our little 🚐 everywhere it goes. Ignore
      the design; the <i>positioning semantics</i> are the important part.

      <SummaryAndHistogram
        min={9508263}
        qlow={21123818}
        median={30627455}
        mean={34851293}
        qhigh={45410890}
        max={52802608}
        width={300}
        height={65}
        data={data01}
        color="black"
      />
    </div>
    <!-- </TooltipContent> -->
  </Tooltip>
</div>
