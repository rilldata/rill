<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createRuntimeServiceListGitCommits } from "@rilldata/web-common/runtime-client";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let referenceCommitSha: string;
  export let onSubmit: () => void;

  const client = useRuntimeClient();
  const listCommitsQuery = createRuntimeServiceListGitCommits(client, {});
  $: commits = $listCommitsQuery.data?.commits ?? [];
  $: referenceCommitIndex = commits.findIndex(
    (c) => c.commitSha === referenceCommitSha,
  );
  $: commitsAfterReference =
    referenceCommitIndex >= 0
      ? commits.slice(0, referenceCommitIndex)
      : commits;
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Title>
      There are changes made after this that will be rolled back.
    </Dialog.Title>
    <Dialog.Description
      class="flex flex-col gap-3 max-h-[800px] overflow-y-auto"
    >
      {#each commitsAfterReference as commit (commit.commitSha)}
        <div>
          <div>{commit.message}</div>
          <div>Author: {commit.authorName ?? commit.authorEmail}</div>
          {#if commit.committedOn}
            <div>Date: {commit.committedOn.toString()}</div>
          {/if}
        </div>
      {/each}
    </Dialog.Description>

    <div class="flex flex-row mt-4 gap-2">
      <div class="grow" />
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        onClick={() => {
          open = false;
          onSubmit();
        }}
        type="primary"
      >
        Save
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
