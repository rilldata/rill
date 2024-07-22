<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceCreateProjectWhitelistedDomain } from "@rilldata/web-admin/client";
  import { getUserDomain } from "@rilldata/web-admin/features/projects/user-invite/selectors";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-invite/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let allowDomain = false;
  $: userDomain = getUserDomain();
  const addToAllowlist = createAdminServiceCreateProjectWhitelistedDomain();
  async function onContinue() {
    if (allowDomain) {
      await $addToAllowlist.mutateAsync({
        organization,
        project,
        data: {
          domain: $userDomain.data,
          role: "viewer",
        },
      });
    }
    return goto(`/${organization}/${project}/-/status`);
  }
</script>

<div
  class="flex flex-col gap-1.5 max-w-[1000px] my-16 sm:my-32 md:my-64 mx-auto"
>
  <div class="text-xl text-center w-full">Invite teammates to your project</div>
  <div>Invite by email</div>
  <UserInviteForm {organization} {project} />
  {#if $userDomain.data}
    <div>Allow domain access</div>
    <div>
      <Switch small bind:checked={allowDomain} id="allow-domain" />
      <Label for="allow-domain">
        Allow any user with a @{$userDomain.data} email address to join this project
        as a Viewer.
        <a
          target="_blank"
          href="https://docs.rilldata.com/reference/cli/user/whitelist/"
        >
          Learn more
        </a>
      </Label>
    </div>
  {/if}
  <Button
    type="primary"
    on:click={onContinue}
    loading={$addToAllowlist.isLoading}>Continue</Button
  >
</div>
