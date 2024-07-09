<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceApproveProjectAccess,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { parseAccessRequestError } from "@rilldata/web-admin/features/access-request/utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import AccessRequestContainer from "@rilldata/web-admin/features/access-request/AccessRequestContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import type { AxiosError } from "axios";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  $: approveAccess = createAdminServiceApproveProjectAccess();
  function onApprove() {
    requested = true;
    void $approveAccess.mutateAsync({
      organization,
      project,
      id,
      data: {},
    });
    goto(`/${organization}/${project}`);
  }

  $: error = parseAccessRequestError(
    $approveAccess.error as unknown as AxiosError<RpcStatus>,
  );
</script>

<AccessRequestContainer>
  <CheckCircle size="40px" className="text-primary-500" />
  <h2 class="text-lg font-normal">Grant access to this project</h2>
  <Button
    type="primary"
    wide
    on:click={onApprove}
    loading={$approveAccess.isLoading}
    disabled={requested}
  >
    Grant access
  </Button>
  {#if error}
    <div class="text-slate-500 text-base">
      {error}
    </div>
  {/if}
</AccessRequestContainer>
