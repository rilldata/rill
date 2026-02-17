<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";
  import ModeToggle from "@rilldata/web-common/layout/ModeToggle.svelte";

  $: ({ instanceId } = $runtime);
  $: projectTitleQuery = useProjectTitle(instanceId);
  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Project";
  $: homeHref = $previewModeStore ? "/home" : "/";
</script>

<header>
  <a href={homeHref} class="flex-shrink-0">
    <Rill />
  </a>

  <ModeToggle />

  <span class="font-medium truncate" style="color: var(--fg-primary)">
    {projectTitle}
  </span>

  <div class="ml-auto flex gap-x-2 h-full w-fit items-center py-2">
    <LocalAvatarButton />
  </div>
</header>

<style lang="postcss">
  header {
    @apply w-full box-border;
    @apply flex gap-x-2 items-center px-4 flex-none;
    @apply h-11;
    background: var(--surface-base);
    border-bottom: 1px solid var(--border);
  }
</style>
