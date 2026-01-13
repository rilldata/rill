<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  const { darkMode } = featureFlags;

  $: ({ instanceId } = $runtime);
  $: projectTitleQuery = useProjectTitle(instanceId);
  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Project";
</script>

<header>
  <a href="/" class="flex-shrink-0">
    <Rill />
  </a>

  <span class="rounded-full px-2 border text-gray-800 bg-gray-50 dark:bg-gray-900 dark:border-gray-800 dark:text-gray-200 text-xs font-medium">
    Developer
  </span>

  <span class="text-gray-900 dark:text-white font-medium truncate">
    {projectTitle}
  </span>

  <div class="ml-auto flex gap-x-2 h-full w-fit items-center py-2">
    <LocalAvatarButton darkMode={$darkMode} />
  </div>
</header>

<style lang="postcss">
  header {
    @apply w-full bg-white dark:bg-gray-950 box-border;
    @apply flex gap-x-2 items-center px-4 flex-none;
    @apply h-11 border-b border-gray-200 dark:border-gray-800;
  }
</style>
