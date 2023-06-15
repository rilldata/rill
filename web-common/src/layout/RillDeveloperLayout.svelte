<script lang="ts">
  import { page } from "$app/stores";
  import NotificationCenter from "@rilldata/web-common/components/notifications/NotificationCenter.svelte";
  import { projectShareStore } from "@rilldata/web-common/features/dashboards/dashboard-stores.js";
  import DeployDashboardOverlay from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardOverlay.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import {
    addReconcilingOverlay,
    syncFileSystemPeriodically,
  } from "@rilldata/web-common/features/entity-management/sync-file-system";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import DuplicateSource from "@rilldata/web-common/features/sources/add-source/DuplicateSource.svelte";
  import FileDrop from "@rilldata/web-common/features/sources/add-source/FileDrop.svelte";
  import { duplicateSourceName } from "@rilldata/web-common/features/sources/sources-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import { initMetrics } from "@rilldata/web-common/metrics/initMetrics";
  import { batcher } from "@rilldata/web-common/runtime-client/http-client";
  import type { ApplicationBuildMetadata } from "@rilldata/web-local/lib/application-state-stores/build-metadata";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";
  import { getArtifactErrors } from "../features/entity-management/getArtifactErrors";
  import PreparingImport from "../features/sources/add-source/PreparingImport.svelte";
  import WelcomePageRedirect from "../features/welcome/WelcomePageRedirect.svelte";
  import { runtimeServiceGetConfig } from "../runtime-client/manual-clients";
  import { runtime } from "../runtime-client/runtime-store";
  import BasicLayout from "./BasicLayout.svelte";
  import { importOverlayVisible, overlay } from "./overlay-store";

  const queryClient = useQueryClient();

  const appBuildMetaStore: Writable<ApplicationBuildMetadata> =
    getContext("rill:app:metadata");

  onMount(async () => {
    const config = await runtimeServiceGetConfig();

    featureFlags.set({
      readOnly: config.readonly,
    });
    batcher.setInstanceId(config.instance_id);

    appBuildMetaStore.set({
      version: config.version,
      commitHash: config.build_commit,
    });

    const res = await getArtifactErrors(config.instance_id);
    fileArtifactsStore.setErrors(res.affectedPaths, res.errors);

    initMetrics(config);
  });

  syncFileSystemPeriodically(
    queryClient,
    runtime,
    featureFlags,
    page,
    fileArtifactsStore
  );
  $: addReconcilingOverlay($page.url.pathname);

  let dbRunState = "disconnected";
  let runstateTimer;

  function debounceRunstate(state) {
    if (runstateTimer) clearTimeout(runstateTimer);
    setTimeout(() => {
      dbRunState = state;
    }, 500);
  }

  let showDropOverlay = false;

  // TODO: add new global run state
  $: debounceRunstate("disconnected");

  function isEventWithFiles(event: DragEvent) {
    let types = event.dataTransfer.types;
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
  {#if $projectShareStore}
    <DeployDashboardOverlay />
  {/if}

  <div
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
