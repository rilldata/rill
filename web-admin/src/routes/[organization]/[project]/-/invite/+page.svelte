<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateProjectWhitelistedDomain,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { showWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import {
    getUserDomain,
    userDomainIsPublic,
  } from "@rilldata/web-admin/features/projects/user-management/selectors";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { AxiosError } from "axios";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ showWelcomeDialog } = data);

  $: if (showWelcomeDialog) {
    showWelcomeToRillDialog.set(true);
  }

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let allowDomain = false;
  let invited = false;
  $: userDomain = getUserDomain();
  $: isPublicDomain = userDomainIsPublic();
  const addToAllowlist = createAdminServiceCreateProjectWhitelistedDomain();

  $: buttonText = invited || allowDomain ? "Continue" : "Skip for now";

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  async function onContinue() {
    if (allowDomain) {
      try {
        await $addToAllowlist.mutateAsync({
          org: organization,
          project,
          data: {
            domain: $userDomain.data,
            role: ProjectUserRoles.Viewer,
          },
        });
      } catch (e) {
        eventBus.emit("notification", {
          type: "error",
          message:
            (e as AxiosError<RpcStatus>).response.data?.message ?? e.message,
          options: {
            persisted: true,
          },
        });
      }
    }
    return goto(getDeployLandingPage());
  }

  function getDeployLandingPage() {
    const u = new URL($page.url);
    u.pathname = `/${organization}/${project}/-/dashboards`;
    u.searchParams.set("deploying", "true");
    return u.toString();
  }
</script>

<div class="flex flex-col gap-5 w-[600px] my-16 sm:my-32 md:my-64 mx-auto">
  <div class="text-xl text-center w-full">Invite teammates to your project</div>
  <div class="flex flex-col gap-y-1">
    <div class="flex flex-row items-center">
      <div class="text-sm font-medium">Invite by email</div>
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
    <UserInviteForm
      {organization}
      {project}
      onInvite={() => (invited = true)}
    />
  </div>
  {#if $userDomain.data && !$isPublicDomain.data}
    <div class="flex flex-col gap-y-1">
      <div class="text-sm font-medium">Allow domain access</div>
      <div class="flex flex-row gap-x-2">
        <Switch
          small
          bind:checked={allowDomain}
          id="allow-domain"
          class="mt-1"
        />
        <Label for="allow-domain" class="font-normal text-gray-700 text-sm">
          Allow existing and new Rill users with a <b>@{$userDomain.data}</b>
          email address to join this project as a <b>Viewer</b>.
          <a
            target="_blank"
            href="https://docs.rilldata.com/reference/cli/user/whitelist"
          >
            Learn more
          </a>
        </Label>
      </div>
    </div>
  {/if}
  <Button
    type="primary"
    onClick={onContinue}
    loading={$addToAllowlist.isPending}
    wide
    class="mx-auto"
  >
    {buttonText}
  </Button>
</div>
