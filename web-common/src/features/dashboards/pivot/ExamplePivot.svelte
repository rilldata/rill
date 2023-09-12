<script lang="ts">
  import { createQuery, useQueryClient } from "@tanstack/svelte-query";
  import Pivot from "./Pivot.svelte";
  import { getBlock } from "../time-dimension-details/util";

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const handlePos = (evt) => {
    pos = evt.detail;
  };

  const fetchData = (block, delay) => async () => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          block: block,
          data: Array.from(Array(50).keys()).map((x) =>
            Array.from(Array(2).keys()).map((y) => `${block[0] + x},${y}`)
          ),
        });
      }, delay);
    });
  };

  $: block = getBlock(50, pos.y0, pos.y1);
  $: rowQuery = createQuery({
    queryKey: ["example-pivot", block[0], block[1]],
    queryFn: fetchData(block, 1000),
  });

  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();
  const handleGetRowHeaderData = (pos) => {
    const block = getBlock(50, pos.y0, pos.y1);
    const cachedBlock = cache.find(["example-pivot", block[0], block[1]])?.state
      ?.data;
    return cachedBlock ?? null;
  };
</script>

<Pivot
  on:pos={handlePos}
  rowHeaderData={$rowQuery.data}
  getRowHeaderData={handleGetRowHeaderData}
/>
