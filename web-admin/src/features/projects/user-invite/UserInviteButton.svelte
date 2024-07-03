<script lang="ts">
  import { page } from "$app/stores";
  import UserInvite from "@rilldata/web-admin/features/projects/user-invite/UserInvite.svelte";
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
    <UserInvite {organization} {project} onInvited={() => (open = false)} />
  </DropdownMenuContent>
</DropdownMenu>
