<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { editorReturnUrl } from "./editor-return-url-store";
  import { previewLocked } from "./preview-locked-store";
  import { previewModeStore } from "./preview-mode-store";
  import { Pencil, Play } from "lucide-svelte";

  $: ({ params } = $page);

  // Derive the current file artifact (if we're on a /files/[...file] route)
  // so we can decide whether Preview should deep-link to a specific dashboard.
  $: filePath = params.file ? `/${params.file}` : undefined;
  $: fileArtifact = filePath
    ? fileArtifacts.getFileArtifact(filePath)
    : undefined;
  $: resourceNameStore = fileArtifact?.resourceName;
  $: resourceKind = $resourceNameStore?.kind as ResourceKind | undefined;
  $: resourceName =
    $resourceNameStore?.name ??
    (filePath ? getNameFromFile(filePath) : undefined);

  $: previewHref = (() => {
    const path = $page.url.pathname;
    // Already on a viz route: enter preview in place rather than navigating.
    if (path.startsWith("/explore/") || path.startsWith("/canvas/")) {
      return path;
    }
    if (resourceKind === ResourceKind.Explore && resourceName) {
      return `/explore/${resourceName}`;
    }
    if (resourceKind === ResourceKind.Canvas && resourceName) {
      return `/canvas/${resourceName}`;
    }
    return "/dashboards";
  })();

  $: returnHref = $editorReturnUrl ?? "/";

  $: inPreviewMode = $previewModeStore;
  $: showReturn = inPreviewMode && !$previewLocked;
  $: showPreview = !inPreviewMode;

  async function enterPreview() {
    previewModeStore.set(true);
    if ($page.url.pathname !== previewHref) {
      await goto(previewHref);
    }
  }

  async function exitPreview() {
    previewModeStore.set(false);
    if ($page.url.pathname !== returnHref) {
      await goto(returnHref);
    }
  }
</script>

{#if showPreview}
  <Tooltip distance={8} location="bottom">
    <Button
      label="Preview"
      type="secondary"
      preload={false}
      compact
      onClick={enterPreview}
    >
      <div class="flex gap-x-1 items-center">
        <Play size={14} />
        Preview
      </div>
    </Button>
    <TooltipContent slot="tooltip-content">
      {#if previewHref === "/dashboards"}
        Open dashboards
      {:else}
        Preview dashboard
      {/if}
    </TooltipContent>
  </Tooltip>
{:else if showReturn}
  <Tooltip distance={8} location="bottom">
    <Button
      label="Return to editor"
      type="secondary"
      preload={false}
      compact
      onClick={exitPreview}
    >
      <div class="flex gap-x-1 items-center">
        <Pencil size={14} />
        Edit
      </div>
    </Button>
    <TooltipContent slot="tooltip-content">Return to editor</TooltipContent>
  </Tooltip>
{/if}
