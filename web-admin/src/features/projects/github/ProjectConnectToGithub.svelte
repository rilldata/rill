<script lang="ts">
  import { createAdminServiceGetGithubUserStatus } from "@rilldata/web-admin/client";
  import ConnectToGithubConfirmDialog from "@rilldata/web-admin/features/projects/github/ConnectToGithubConfirmDialog.svelte";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  const githubUserStatus = createAdminServiceGetGithubUserStatus();

  let confirmDialogOpen = false;
  let githubSelectionOpen = false;
</script>

<span>
  Unlock the power of BI-as-code with Github-backed collaboration, version
  control, and approval workflows.
</span>
<Button
  type="primary"
  class="w-fit mt-1"
  loading={$githubUserStatus.isFetching}
  on:click={() => (confirmDialogOpen = true)}
>
  Connect to Github
</Button>

<ConnectToGithubConfirmDialog
  bind:open={confirmDialogOpen}
  onContinue={() => (githubSelectionOpen = true)}
/>

<GithubRepoSelectionDialog
  bind:open={githubSelectionOpen}
  onConnect={() => {}}
/>
