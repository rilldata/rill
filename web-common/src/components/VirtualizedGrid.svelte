<script lang="ts">
  /** This is adapted from
   * https://github.com/sveltejs/svelte-virtual-list/blob/master/VirtualList.svelte
   */
  import { onMount, tick } from "svelte";

  // props
  export let items;
  export let height = "100%";
  export let itemHeight = undefined;
  export let columns = 3;
  // read-only, but visible to consumers via bind:start
  export let start = 0;
  export let end = 0;
  export let columnSize: string | "auto" = "auto";
  // local state
  let height_map = [];
  let rows;
  let viewport;
  let contents;
  let viewport_height = 0;
  let visible;
  let mounted;
  let top = 0;
  let bottom = 0;
  let average_height;
  /** The ultimate goal of visible is to slice the available items
   * by start & end indices.
   */
  $: visible = items.slice(start, end).map((data, i) => {
    return { index: i + start, data };
  });
  // whenever `items` changes, invalidate the current heightmap
  $: if (mounted) refresh(items, viewport_height, itemHeight, columns);
  /** goal is to set the start index and update the height_map. */
  async function refresh(items, viewport_height, itemHeight, columns = 3) {
    const { scrollTop } = viewport;
    await tick(); // wait until the DOM is up to date
    let content_height = top - scrollTop;
    let i = start;
    while (content_height < viewport_height && i < items.length) {
      let row = rows[i - start];
      if (!row) {
        end = i + columns;
        await tick(); // render the newly visible row
        row = rows[i - start];
      }
      const row_height = (height_map[i] = itemHeight || row?.offsetHeight) + 28;
      // add this row height
      content_height += row_height;
      i += columns;
    }
    end = i;
    const remaining = items.length - end;
    average_height = (top + content_height) / end;
    bottom = remaining * average_height;
    height_map.length = items.length;
  }
  async function handle_scroll() {
    rows = contents.getElementsByTagName("svelte-virtual-list-row");
    const { scrollTop } = viewport;
    const old_start = start;
    // build a height map of the individual elements in rows.
    for (let v = 0; v < rows.length; v += 1) {
      height_map[start + v] = rows[v].offsetHeight;
    }
    /** What is i?
     *
     */
    let i = 0;
    /** y is the top of the next element. */
    let y = 0;
    /** What does this do?
     *
     */
    while (i < items.length) {
      /** element_height is this specific element height. */
      const element_height = height_map[i] || average_height;
      if (y + element_height > scrollTop) {
        start = i;
        top = y;
        break;
      }
      y += element_height + 28;
      i += columns;
    }
    while (i < items.length) {
      y += height_map[i] || average_height;
      i += columns;
      if (y > scrollTop + viewport_height) break;
    }
    end = i;
    const remaining = items.length - end;
    average_height = y / end;
    /** What's the point of this line?*/
    while (i < items.length) height_map[i++] = average_height;
    bottom = remaining * average_height;
    // prevent jumping if we scrolled up into unknown territory
    if (start < old_start) {
      await tick();
      let expected_height = 0;
      let actual_height = 0;
      for (let i = start; i < old_start; i += columns) {
        if (rows[i - start]) {
          expected_height += height_map[i] + 28;
          actual_height += itemHeight || rows[i - start].offsetHeight + 28;
        }
      }
      const d = actual_height - expected_height;
      viewport.scrollTo(0, scrollTop + d);
    }
    // TODO if we overestimated the space these
    // rows would occupy we may need to add some
    // more. maybe we can just call handle_scroll again?
  }
  // trigger initial refresh
  onMount(() => {
    rows = contents.getElementsByTagName("svelte-virtual-list-row");
    mounted = true;
  });
</script>

<svelte-virtual-list-viewport
  bind:this={viewport}
  bind:offsetHeight={viewport_height}
  on:scroll={handle_scroll}
  style="height: {height};"
  style:width="100%"
>
  <svelte-virtual-list-contents
    bind:this={contents}
    style="
      display: grid;
      width: 100%;
      gap: 2rem;
      justify-items:start;
      justify-content: start;
      grid-template-columns: repeat({columns}, {columnSize});
      padding-top: {top}px;
      padding-bottom: {bottom}px;"
  >
    {#each visible as row (row.index)}
      <svelte-virtual-list-row>
        <slot item={row.data}>Missing template</slot>
      </svelte-virtual-list-row>
    {/each}
  </svelte-virtual-list-contents>
</svelte-virtual-list-viewport>

<style>
  svelte-virtual-list-viewport {
    position: relative;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    display: block;
  }
  svelte-virtual-list-contents,
  svelte-virtual-list-row {
    display: block;
  }
  svelte-virtual-list-row {
    overflow: hidden;
  }
</style>
