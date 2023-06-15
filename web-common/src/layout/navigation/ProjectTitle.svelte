<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { shorthandTitle } from "@rilldata/web-common/layout/navigation/shorthand-title/index.js";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import { parseDocument } from "yaml";

  let mounted = false;
  onMount(() => {
    mounted = true;
  });

  $: projectTitle = createRuntimeServiceGetFile(
    $runtime?.instanceId,
    "rill.yaml",
    {
      query: {
        select: (data) => {
          const projectData = parseDocument(data.blob)?.toJS();
          return projectData.title ?? projectData.name;
        },
      },
    }
  );
</script>

<header
  class="sticky top-0 grid align-center bg-white z-50"
  style:height="var(--header-height)"
>
  <!-- the pl-[.875rem] is a fix to move this new element over a pinch.-->
  <h1
    class="grid grid-flow-col justify-start gap-x-3 p-4 pl-[.75rem] items-center content-center"
  >
    {#if mounted && $projectTitle.isSuccess}
      <a href="/">
        <div
          style:width="20px"
          style:font-size="9px"
          class="grid place-items-center rounded bg-gray-800 text-white font-normal"
          style:height="20px"
        >
          <div>
            {shorthandTitle($projectTitle.data || "Ri")}
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
        {$projectTitle.data || "Untitled Rill Project"}
      </a>
      <TooltipContent maxWidth="300px" slot="tooltip-content">
        <div class="font-bold">
          {$projectTitle.data || "Untitled Rill Project"}
        </div>
      </TooltipContent>
    </Tooltip>
  </h1>
</header>
