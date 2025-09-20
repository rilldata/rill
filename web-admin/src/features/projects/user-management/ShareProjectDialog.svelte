<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import ShareProjectForm from "@rilldata/web-admin/features/projects/user-management/ShareProjectForm.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";

  export let organization: string;
  export let project: string;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;
  export let open = false;

  $: projectQuery = createAdminServiceGetProject(organization, project);
  $: manageProjectAdmins =
    !!$projectQuery.data?.projectPermissions?.manageProjectAdmins;
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="p-0 m-0">
    <ShareProjectForm
      {organization}
      {project}
      {manageProjectAdmins}
      {manageOrgAdmins}
      {manageOrgMembers}
      enabled={open}
    />
  </Dialog.Content>
</Dialog.Root>
