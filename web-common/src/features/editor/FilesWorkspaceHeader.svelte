<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import HideRightSidebar from "@rilldata/web-common/components/icons/HideRightSidebar.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { createResizeListenerActionFactory } from "@rilldata/web-common/lib/actions/create-resize-listener-factory";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { getContext } from "svelte";
  import type { Tweened } from "svelte/motion";
  import type { Writable } from "svelte/store";
  import { WorkspaceHeaderStatusSpinner } from "../../layout/workspace";
  import { EntityStatus } from "../entity-management/types";
  import Breadcrumbs from "./Breadcrumbs.svelte";

  export let filePath: string;

  const { listenToNodeResize } = createResizeListenerActionFactory();

  const outputLayout = getContext(
    "rill:app:output-layout"
  ) as Writable<LayoutElement>;

  const navigationVisibilityTween = getContext(
    "rill:app:navigation-visibility-tween"
  ) as Tweened<number>;

  const inspectorLayout = getContext(
    "rill:app:inspector-layout"
  ) as Writable<LayoutElement>;
</script>

<header
  class="grid items-center content-stretch justify-between pl-4 border-b border-gray-300"
  style:grid-template-columns="[title] auto [controls] auto"
  style:height="var(--header-height)"
  use:listenToNodeResize
>
  <div style:padding-left="{$navigationVisibilityTween * 24}px">
    <Breadcrumbs {filePath} />
  </div>

  <div class="flex items-center mr-4">
    <IconButton
      on:click={() => {
        outputLayout.update((state) => {
          state.visible = !state.visible;
          return state;
        });
      }}
      ><span class="text-gray-500"><HideBottomPane size="18px" /></span>
      <svelte:fragment slot="tooltip-content">
        <SlidingWords active={$outputLayout?.visible} reverse
          >results preview</SlidingWords
        >
      </svelte:fragment>
    </IconButton>

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

    <div class="ml-2">
      <WorkspaceHeaderStatusSpinner
        applicationStatus={$appQueryStatusStore
          ? EntityStatus.Running
          : EntityStatus.Idle}
      />
    </div>
  </div>
</header>
