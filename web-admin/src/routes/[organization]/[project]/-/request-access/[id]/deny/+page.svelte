<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceDenyProjectAccess,
    createAdminServiceGetProjectAccess,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseAccessRequestError } from "@rilldata/web-admin/features/access-request/utils";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  $: denyAccess = createAdminServiceDenyProjectAccess();
  $: if (organization && project && id && !requested) {
    requested = true;
    void $denyAccess.mutateAsync({
      organization,
      project,
      id,
      data: {},
    });
  }
  $: error = parseAccessRequestError(
    $denyAccess.error as unknown as AxiosError<RpcStatus>,
  );

  $: requestAccess = createAdminServiceGetProjectAccess(
    organization,
    project,
    id,
  );
</script>

<AccessRequestContainer>
  {#if $denyAccess.isLoading}
    <div class="h-36 mt-10">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
  {:else if error}
    <div class="text-slate-500 text-base">
      {error}
    </div>
  {:else}
    <CrossIcon size="30px" className="text-red-500" />
    <h2 class="text-lg font-normal">Access denied</h2>
    {#if $requestAccess.data}
      <div class="text-slate-500 text-base">
        <b>{$requestAccess.data.email}</b> denied access to <b>{project}</b>.
      </div>
    {/if}
  {/if}
</AccessRequestContainer>
