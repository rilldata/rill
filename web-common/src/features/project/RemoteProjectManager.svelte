<script lang="ts">
  import { page } from "$app/stores";
  import RemoteSyncDialogs from "@rilldata/web-common/features/project/RemoteSyncDialogs.svelte";
  import { createRuntimeServiceGitStatus } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { derived } from "svelte/store";

  const runtimeClient = useRuntimeClient();
  const gitStatusQuery = createRuntimeServiceGitStatus(runtimeClient, {});

  const gitStatusSource = derived(gitStatusQuery, ($q) => ({
    hasRemoteChanges: Boolean($q.data?.remoteCommits),
    hasLocalCommitsOnCurrent: Boolean($q.data?.localCommits),
  }));

  $: inDeployPage = $page.route.id?.startsWith("/(misc)/deploy") ?? false;
</script>

<RemoteSyncDialogs {gitStatusSource} autoOpen={!inDeployPage} />
