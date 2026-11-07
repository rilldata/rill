<script lang="ts">
  import { page } from "$app/state";
  import CanvasPersonalFile from "@rilldata/web-admin/features/personal-files/canvas/CanvasPersonalFile.svelte";
  import { getAdminServiceGetPersonalFileQueryKey } from "@rilldata/web-admin/client";
  import { adminServiceGetPersonalFileUsingToken } from "@rilldata/web-admin/features/personal-files/personal-file-client-using-token.ts";
  import { VirtualFileIo } from "@rilldata/web-admin/features/personal-files/virtual-file-io.ts";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery } from "@tanstack/svelte-query";

  const client = useRuntimeClient();
  const adminToken = EmbedStore.getInstance()?.adminToken ?? "";

  // Embeds don't have the org and project in the URL, so we resolve them from the runtime instance.
  const orgAndProjectQuery = getOrgAndProjectForEmbed(client);
  let { organization, project } = $derived($orgAndProjectQuery);

  let name = $derived(page.params.name ?? "");

  let fileIo = $derived(
    organization && project
      ? new VirtualFileIo(organization, project, adminToken)
      : undefined,
  );

  // Route file reads and writes for the personal canvas through the admin server, which stores it as a virtual file.
  let fileIoReady = $state(false);
  $effect(() => {
    if (!fileIo) return;
    fileArtifacts.setClient(client, fileIo);
    fileIoReady = true;

    return fileIo.on("write", (args) => void onFileWrite(args));
  });

  async function onFileWrite({ name, kind }: { name: string; kind: string }) {
    // Invalidate the resource caches so the preview re-fetches the reconciled canvas
    // (otherwise the user might see a stale render right after their edit).
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGetResourceQueryKey(client.instanceId, {
        name: { kind, name },
      }),
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(client.instanceId, {}),
    });
  }

  let personalFileQuery = $derived(
    createQuery(
      {
        queryKey: getAdminServiceGetPersonalFileQueryKey(
          organization,
          project,
          name,
        ),
        queryFn: ({ signal }) =>
          adminServiceGetPersonalFileUsingToken(
            organization,
            project,
            name,
            adminToken,
            signal,
          ),
        enabled: !!organization && !!project && !!name,
        // The file IO keeps the cached file up to date on writes, so refetching could only clobber in-flight edits.
        staleTime: Infinity,
      },
      queryClient,
    ),
  );
  let personalFile = $derived($personalFileQuery.data);

  let fileArtifact = $derived(
    personalFile
      ? fileArtifacts.getFileArtifact(addLeadingSlash(personalFile.path ?? ""))
      : undefined,
  );
  $effect(() => {
    if (fileArtifact && personalFile) {
      fileArtifact.editorContent.set(personalFile.yaml ?? "");
    }
  });
</script>

{#if $personalFileQuery.isError}
  <ErrorPage header={m.embed_personal_canvas_load_error()} />
{:else if fileIo && fileIoReady && fileArtifact}
  {#key name}
    <CanvasPersonalFile
      {fileArtifact}
      {fileIo}
      {organization}
      {project}
      {name}
    />
  {/key}
{:else}
  <div class="flex items-center justify-center h-full">
    <DelayedSpinner isLoading={true} size="24px" />
  </div>
{/if}
