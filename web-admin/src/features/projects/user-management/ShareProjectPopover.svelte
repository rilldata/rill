<script lang="ts">
  import ShareProjectForm from "@rilldata/web-admin/features/projects/user-management/ShareProjectForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";

  export let organization: string;
  export let project: string;
  export let manageProjectAdmins: boolean;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;

  let open = false;
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary" selected={open}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[520px]" padding="0">
    <ShareProjectForm
      {organization}
      {project}
      {manageProjectAdmins}
      {manageOrgAdmins}
      {manageOrgMembers}
      enabled={open}
    />
  </PopoverContent>
</Popover>
