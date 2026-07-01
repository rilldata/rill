<script lang="ts">
  import { goto } from "$app/navigation";
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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
      if (rpcError.code === 9) {
        // FailedPrecondition: the user already has access, so send them to the project.
        void goto(`/${organization}/${project}`);
      } else if (rpcError.code !== 6) {
        // do not show error if already requested invite (AlreadyExists)
        errorMessage = rpcError.message;
      }
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
    className={requested ? "text-fg-secondary" : "text-primary-600"}
  />
  <h2 class="text-lg font-normal">{m.auth_request_access_title()}</h2>
  <div class="text-fg-secondary text-base">
    {m.auth_request_access_description({ project })}
  </div>
  <Button
    type="primary"
    wide
    onClick={onRequestAccess}
    loading={isPending}
    disabled={requested}
  >
    {#if requested}<Check />{m.auth_access_requested()}{:else}{m.auth_request_access()}{/if}
  </Button>
  {#if requested && !isPending}
    {#if errorMessage}
      <div>{errorMessage}</div>
    {:else}
      <div class="text-fg-secondary">
        {m.auth_request_sent()}
      </div>
    {/if}
  {/if}
</AccessRequestContainer>
