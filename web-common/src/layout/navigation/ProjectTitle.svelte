<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { shorthandTitle } from "@rilldata/web-common/layout/navigation/shorthand-title/index.js";
  import { useLocalProjectTitle } from "@rilldata/web-common/layout/navigation/useLocalProjectTitle";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: title = useLocalProjectTitle($runtime?.instanceId);
</script>

<header
  class="sticky top-0 grid align-center bg-white z-50"
  style:height="var(--header-height)"
>
  <!-- the pl-[.875rem] is a fix to move this new element over a pinch.-->
  <h1
    class="grid grid-flow-col justify-start gap-x-3 p-4 pl-[.75rem] items-center content-center"
  >
    {#if mounted}
      <a href="/">
        <div
          style:width="20px"
          style:font-size="9px"
          class="grid place-items-center rounded bg-gray-800 text-white font-normal"
          style:height="20px"
        >
          <div>
            {shorthandTitle(title || "Ri")}
          </div>
        </div>
      </a>
    {:else}
      <Spacer size="16px" />
    {/if}
    <Tooltip distance={8}>
      <a
        class="font-semibold text-black grow text-ellipsis overflow-hidden whitespace-nowrap pr-12"
        href="/"
      >
        {title || "Untitled Rill Project"}
      </a>
      <TooltipContent maxWidth="300px" slot="tooltip-content">
        <div class="font-bold">
          {title || "Untitled Rill Project"}
        </div>
      </TooltipContent>
    </Tooltip>
  </h1>
</header>
