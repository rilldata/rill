<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRequestProjectAccess,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import type { AxiosError } from "axios";

  $: organization = $page.url.searchParams.get("organization");
  $: project = $page.url.searchParams.get("project");
  $: role = $page.url.searchParams.get("role") ?? ProjectUserRoles.Viewer;
  $: autoRequest = $page.url.searchParams.get("auto_request") === "true";
  $: if (autoRequest) onRequestAccess();

  let requested = false;
  const requestAccess = createAdminServiceRequestProjectAccess();

  let errorMessage = "";
  $: if ($requestAccess.error) {
    const rpcError = ($requestAccess.error as unknown as AxiosError<RpcStatus>)
      .response.data;
    if (rpcError) {
      // do not show error if already requested invite
      if (rpcError.code !== 6) errorMessage = rpcError.message;
    } else {
      errorMessage = $requestAccess.error.toString();
    }
  }

  $: isPending = $requestAccess.isPending;

  function onRequestAccess() {
    requested = true;
    void $requestAccess.mutateAsync({
      org: organization,
      project,
      data: {
        role,
      },
    });
  }
</script>

<AccessRequestContainer>
  <Lock
    size="40px"
    className={requested ? "text-gray-600" : "text-primary-600"}
  />
  <h2 class="text-lg font-normal">Request access to this project</h2>
  <div class="text-slate-500 text-base">
    You can view <b>{project}</b> once your request is approved.
  </div>
  <Button
    type="primary"
    wide
    onClick={onRequestAccess}
    loading={isPending}
    disabled={requested}
  >
    {#if requested}<Check />Access requested{:else}Request access{/if}
  </Button>
  {#if requested && !isPending}
    {#if errorMessage}
      <div>{errorMessage}</div>
    {:else}
      <div class="text-slate-500">
        Your request has been sent to the project admin. You’ll get an email
        when it’s approved.
      </div>
    {/if}
  {/if}
</AccessRequestContainer>
