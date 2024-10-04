<script lang="ts">
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import AddDataModal from "@rilldata/web-common/features/sources/modal/AddDataModal.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Play from "svelte-radix/Play.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Pencil from "svelte-radix/Pencil1.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";
  import { get } from "svelte/store";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";

  let showDropOverlay = false;

  $: ({
    url: { pathname },
  } = $page);

  $: ({ instanceId } = $runtime);

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }

  $: projectTitleQuery = useProjectTitle(instanceId);

  $: ({ data: title } = $projectTitleQuery);

  async function submitTitleChange(editedTitle: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");

    let content = get(artifact.localContent) ?? get(artifact.remoteContent);

    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.localContent) ?? get(artifact.remoteContent);
      if (!content) {
        return;
      }
    }
    const parsed = parseDocument(content);

    parsed.set("title", editedTitle);

    artifact.updateLocalContent(parsed.toString(), true);
    await artifact.saveLocalContent();
  }
</script>

<main
  role="application"
  class="index-body relative size-full flex flex-col overflow-hidden"
  on:drag|preventDefault|stopPropagation
  on:drop|preventDefault|stopPropagation
  on:dragenter|preventDefault|stopPropagation
  on:dragleave|preventDefault|stopPropagation
  on:dragover|preventDefault|stopPropagation={(e) => {
    if (isEventWithFiles(e)) showDropOverlay = true;
  }}
>
  {#if pathname !== "/welcome"}
    <header>
      <a href="/">
        <Rill />
      </a>

      <span class="rounded-full px-2 border text-gray-800 bg-gray-50">
        Developer
      </span>

      <div class="h-6 w-fit ml-2">
        <InputWithConfirm value={title} onConfirm={submitTitleChange} />
      </div>

      <div class="ml-auto flex gap-x-2">
        <Button square type="secondary">
          <Play size="16px" />
        </Button>
        <DeployDashboardCta />

        <LocalAvatarButton />
      </div>
    </header>
  {/if}
  <div class="flex size-full overflow-hidden">
    {#if pathname !== "/welcome"}
      <Navigation />
    {/if}
    <section class="size-full overflow-hidden">
      <slot />
    </section>
  </div>
</main>

{#if showDropOverlay}
  <FileDrop bind:showDropOverlay />
{:else if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

<AddDataModal />
<SourceImportedModal sourcePath={$sourceImportedPath} />

<style lang="postcss">
  header {
    @apply w-full bg-background box-border;
    @apply flex gap-x-2 items-center px-4 border-b flex-none;
    height: var(--header-height);
  }
</style>
