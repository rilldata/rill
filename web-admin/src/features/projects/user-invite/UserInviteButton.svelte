<script lang="ts">
  import { page } from "$app/stores";
  import UserInviteAllowlist from "@rilldata/web-admin/features/projects/user-invite/UserInviteAllowlist.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-invite/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
    DropdownMenuContent,
  } from "@rilldata/web-common/components/dropdown-menu";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  let open = false;
</script>

<DropdownMenu bind:open>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary">Share</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent class="w-[520px] p-4">
    <div class="flex flex-col gap-1.5">
      <div class="text-base font-medium">Share this project</div>
      <UserInviteForm
        {organization}
        {project}
        onInvite={() => (open = false)}
      />
      <UserInviteAllowlist {organization} {project} />
    </div>
  </DropdownMenuContent>
</DropdownMenu>
