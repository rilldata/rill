<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { shorthandTitle } from "@rilldata/web-common/features/project/shorthand-title/index.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useProjectTitle } from "./selectors";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";

  export let unsavedFileCount: number;

  $: ({ instanceId } = $runtime);

  $: projectTitle = useProjectTitle(instanceId);

  $: ({ data: title } = $projectTitle);
</script>

<header class="pl-3 items-center flex flex-none">
  <div class="link-wrapper">
    <a href="/" class="project-square">
      {shorthandTitle(title ?? "Rill")}

      {#if unsavedFileCount > 0}
        <Tooltip distance={8} location="right">
          <div class="unsaved-indicator">
            {unsavedFileCount}
          </div>
          <TooltipContent maxWidth="300px" slot="tooltip-content">
            <TooltipTitle>
              <svelte:fragment slot="name">
                {unsavedFileCount} unsaved file{#if unsavedFileCount > 1}s{/if}
              </svelte:fragment>
            </TooltipTitle>

            <TooltipShortcutContainer>
              Save all
              <Shortcut>
                <MetaKey withAlt action="S" />
              </Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      {/if}
    </a>

    {#if title}
      <Tooltip distance={8}>
        <a class="project-link" href="/">
          {title}
        </a>

        <TooltipContent maxWidth="300px" slot="tooltip-content">
          Go to home page
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</header>

<style lang="postcss">
  header {
    height: var(--header-height);
  }

  .project-square {
    @apply relative;
    @apply h-5 aspect-square grid place-items-center rounded;
    @apply bg-gray-800 text-white font-normal;
    font-size: 9px;
  }

  .link-wrapper {
    @apply flex gap-x-3 items-center flex-none w-full;
  }

  .project-link {
    @apply text-black font-semibold truncate pr-9;
  }

  .project-link:hover {
    @apply text-primary-600;
  }

  .unsaved-indicator {
    @apply -top-1.5 -right-1.5 absolute bg-primary-500 text-white text-[9px] w-fit min-w-[14px] px-1 items-center justify-center flex h-[14px] rounded-full flex-none;
  }
</style>
