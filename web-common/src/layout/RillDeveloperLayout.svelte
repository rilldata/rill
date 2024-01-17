<script lang="ts">
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { resourcesStore } from "@rilldata/web-common/features/entity-management/resources-store";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import DuplicateSource from "@rilldata/web-common/features/sources/modal/DuplicateSource.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/modal/FileDrop.svelte";
  import SourceImportedModal from "@rilldata/web-common/features/sources/modal/SourceImportedModal.svelte";
  import {
    duplicateSourceName,
    sourceImportedName,
  } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import type { ApplicationBuildMetadata } from "@rilldata/web-common/layout/build-metadata";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import PreparingImport from "../features/sources/modal/PreparingImport.svelte";
  import WelcomePageRedirect from "../features/welcome/WelcomePageRedirect.svelte";
  import { runtimeServiceGetConfig } from "../runtime-client/manual-clients";
  import BasicLayout from "./BasicLayout.svelte";
  import { importOverlayVisible, overlay } from "./overlay-store";

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  onMount(async () => {
    const config = await runtimeServiceGetConfig();
    initMetrics(config);

    featureFlags.set({
      adminServer: false,
      readOnly: config.readonly,
    });

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });

    return resourcesStore.init(config.instance_id);
  });

  let showDropOverlay = false;

  function isEventWithFiles(event: DragEvent) {
    let types = event?.dataTransfer?.types;
    return types && types.indexOf("Files") != -1;
  }
</script>

<div class="body">
  {#if $importOverlayVisible}
    <PreparingImport />
  {:else if showDropOverlay}
    <FileDrop bind:showDropOverlay />
  {:else if $overlay !== null}
    <BlockingOverlayContainer
      bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
    >
      <div slot="title">
        <span class="font-bold">{$overlay?.title}</span>
        {#if $overlay?.message}
          <div>{$overlay?.message}</div>
        {/if}
      </div>
    </BlockingOverlayContainer>
  {/if}

  {#if $duplicateSourceName !== null}
    <DuplicateSource />
  {/if}
  <SourceImportedModal open={!!$sourceImportedName} />

  <div
    role="application"
    class="index-body absolute w-screen h-screen"
    on:dragenter|preventDefault|stopPropagation
    on:dragleave|preventDefault|stopPropagation
    on:dragover|preventDefault|stopPropagation={(e) => {
      if (isEventWithFiles(e)) showDropOverlay = true;
    }}
    on:drag|preventDefault|stopPropagation
    on:drop|preventDefault|stopPropagation
  >
    <WelcomePageRedirect>
      <BasicLayout>
        <slot />
      </BasicLayout>
    </WelcomePageRedirect>
  </div>
</div>

<NotificationCenter />

<style>
  /* Prevent trackpad navigation (like other code editors, like vscode.dev). */
  :global(body) {
    overscroll-behavior: none;
  }
</style>
