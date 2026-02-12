<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceRedeployProject,
  } from "@rilldata/web-admin/client";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/github/ProjectGithubConnection.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected =
    !!$proj.data?.project?.gitRemote && !$proj.data?.project?.managedGitId;

  let isRedeployDialogOpen = false;
  const redeployProject = createAdminServiceRedeployProject();

  async function handleRedeploy() {
    await $redeployProject.mutateAsync({
      org: organization,
      project,
    });
    isRedeployDialogOpen = false;
  }
</script>

<div class="flex flex-col gap-y-4 size-full">
  <SettingsContainer title="GitHub" suppressFooter={isGithubConnected}>
    <div slot="body">
      <ProjectGithubConnection {organization} {project} />
    </div>
    <div slot="action">
      <GithubConnectionDialog {organization} {project} />
    </div>
  </SettingsContainer>

  <SettingsContainer title="Redeploy">
    <div slot="body">
      <p>
        Create a new deployment from scratch. This will re-ingest all data and
        may take some time. Use this if your deployment is in a bad state or
        you need a clean slate.
      </p>
    </div>
    <div slot="action">
      <Button
        type="destructive"
        onClick={() => {
          isRedeployDialogOpen = true;
        }}
      >
        Redeploy project
      </Button>
    </div>
  </SettingsContainer>
</div>

<AlertDialog bind:open={isRedeployDialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Redeploy project?</AlertDialogTitle>
      <AlertDialogDescription>
        This will create a completely new deployment and re-ingest all data.
        The existing deployment will be deprovisioned. This may take some time.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          isRedeployDialogOpen = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={handleRedeploy}>Yes, redeploy</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
