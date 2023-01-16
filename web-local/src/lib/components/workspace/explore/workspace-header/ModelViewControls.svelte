<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";

  const inspectorLayout = getContext(
    "rill:app:inspector-layout"
  ) as Writable<LayoutElement>;
</script>

<IconButton
  on:click={() => {
    inspectorLayout.update((state) => {
      state.visible = !state.visible;
      return state;
    });
  }}
>
  <span class="text-gray-500">
    <HideRightSidebar size="18px" />
  </span>
  <svelte:fragment slot="tooltip-content">
    <SlidingWords
      active={$inspectorLayout?.visible}
      direction="horizontal"
      reverse>inspector</SlidingWords
    >
  </svelte:fragment>
</IconButton>
