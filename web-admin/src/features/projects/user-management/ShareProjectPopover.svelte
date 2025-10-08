<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ShareProjectForm from "@rilldata/web-admin/features/projects/user-management/ShareProjectForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils.ts";
  import { onMount } from "svelte";

  export let organization: string;
  export let project: string;
  export let manageProjectAdmins: boolean;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;

  let open = false;

  onMount(() => {
    if ($page.url.searchParams.get("share") === "true") {
      // If we are showing the share popover directly, then unset the param from the url.
      // This prevents the user from saving/sharing a url that would open the share popover.
      void goto(copyWithAdditionalArguments($page.url, {}, { share: false }), {
        replaceState: true,
      });
      open = true;
    }
  });
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
