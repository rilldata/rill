<script lang="ts">
  import { onMount } from "svelte";

  // - desired behavior:
  //   - if both tables fit in the available space, don’t shrink/expand either
  //   - if the tables exceed the available height, and one takes up less than 50% of the space, the one with more rows should shrink
  //   - if the tables exceed the available height and both take up more than 50% of the available space, both should have max height 50%

  // const containers = { outer: undefined, top: undefined, bottom: undefined };
  // const heights = { outer: 0, top: 0, bottom: 0 };

  let topPxRequested: number = undefined;
  const topResizeCallback = (pxRequested: number) => {
    topPxRequested = pxRequested;
  };

  let bottomPxRequested: number = undefined;
  const bottomResizeCallback = (pxRequested: number) => {
    bottomPxRequested = pxRequested;
  };

  let container;
  let containerHeight;
  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerHeight = container.clientHeight;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });

  let topStyle = "";
  let bottomStyle = "";
  $: {
    if (topPxRequested + bottomPxRequested < containerHeight) {
      // if they fit within the available space, no style needs to be applied
      console.log("height ok");
      topStyle = "";
      bottomStyle = "";
    } else if (
      topPxRequested > containerHeight / 2 &&
      bottomPxRequested > containerHeight / 2
    ) {
      // if both are taller than 50% of the container, limit both to 50%
      console.log("both big");
      topStyle = "height:50%;";
      bottomStyle = "height:50%;";
    } else {
      // in total, the elements are taller than the available height, but only one
      // exceed 50% of height; just shrink the taller one to the remaining space
      const BUFFER_PX = 0;
      if (topPxRequested > containerHeight / 2) {
        // top is biggger, so give bottom all the space it wants +
        topStyle = `height:${
          containerHeight - bottomPxRequested - BUFFER_PX
        }px`;
        bottomStyle = `height:${bottomPxRequested + BUFFER_PX}px`;
      } else {
        topStyle = `height:${topPxRequested + BUFFER_PX}px`;
        bottomStyle = `height:${
          containerHeight - topPxRequested - BUFFER_PX
        }px`;
      }
    }
  }
</script>

<div style:height="100%" bind:this={container}>
  <div style={topStyle}>
    <slot name="top-item" {topResizeCallback} />
  </div>
  <div style={bottomStyle}>
    <slot name="bottom-item" {bottomResizeCallback} />
  </div>
</div>

<!-- 
<LayoutManager let:topResizeCallback let:bottomResizeCallback >
  <table slot="top-item" resizeCallback={topResizeCallback}/>
  <table slot="bottom-item" resizeCallback={bottomResizeCallback} />
</LayoutManager> -->

<!-- 
  <script lang="ts">
  import { onMount } from "svelte";

  const containers = {
    outer: undefined,
    topHeader: undefined,
    topTable: undefined,
    bottomHeader: undefined,
    bottomTable: undefined,
  };
  const heights = { outer: 0, top: 0, bottom: 0 };

  onMount(() => {
    const observerAndContainers = Object.entries(containers).map(
      ([k, container]) => {
        const observer = new ResizeObserver(() => {
          heights[k] = container.clientHeight;
        });
        observer.observe(container);
        return [observer, container] as const;
      }
    );
    return () => {
      observerAndContainers.forEach(([observer, container]) => {
        observer.unobserve(container);
      });
    };
  });

  // $: {
  //   if ()
  // }
</script>

<div bind:this={containers.outer}>
  <div bind:this={containers.topHeader}>
    <slot name="top-header" />
  </div>

  <div bind:this={containers.topTable}>
    <slot name="top-table" />
  </div>

  <div bind:this={containers.bottomTable}>
    <slot name="bottom-table" />
  </div>

  <div bind:this={containers.bottomHeader}>
    <slot name="bottom-header" />
  </div>
</div>

 -->
