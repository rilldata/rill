<script lang="ts">
  import { VirtualFileIo } from "@rilldata/web-admin/features/virtual-file-editor/virtual-file-io.ts";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import VirtualCanvasEditor from "@rilldata/web-admin/features/virtual-file-editor/canvas/VirtualCanvasEditor.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

  let {
    name,
    kind,
    org,
    project,
  }: {
    name: string;
    kind: ResourceKind;
    org: string;
    project: string;
  } = $props();

  const client = useRuntimeClient();

  const user = createAdminServiceGetCurrentUser();
  let userId = $derived($user.data?.user?.id ?? "");

  let path = $derived(`/personal/${name}_${userId}`);

  let fileIO = $derived(new VirtualFileIo(org, project, userId, onFileWrite));
  let fileArtifact = $derived(new FileArtifact(client, path, fileIO));

  async function onFileWrite(name: string, kind?: string) {
    if (!kind) return;
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
</script>

{#if kind === ResourceKind.Canvas}
  <VirtualCanvasEditor {fileArtifact} />
{:else}
  Unsupported resource kind: {kind}
{/if}
