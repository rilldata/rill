<script lang="ts">
  import { page } from "$app/stores";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-invite/CopyInviteLinkButton.svelte";
  import UserInviteAllowlist from "@rilldata/web-admin/features/projects/user-invite/UserInviteAllowlist.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-invite/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
    DropdownMenuContent,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let organization: string;
  export let project: string;
  let open = false;

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;
</script>

<DropdownMenu bind:open>
  <DropdownMenuTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary">Share</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent class="w-[520px] p-4" side="bottom" align="end">
    <div class="flex flex-col gap-y-3">
      <div class="flex flex-row items-center">
        <div class="text-sm font-medium">Share this project</div>
        <div class="grow"></div>
        <CopyInviteLinkButton {copyLink} />
      </div>
      <UserInviteForm
        {organization}
        {project}
        onInvite={() => (open = false)}
      />
      <UserInviteAllowlist {organization} {project} />
    </div>
  </DropdownMenuContent>
</DropdownMenu>
