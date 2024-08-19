<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    type V1Quotas,
  } from "@rilldata/web-admin/client";
  import { formatDataSizeQuota } from "@rilldata/web-admin/features/billing/plans/utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { Progress } from "@rilldata/web-common/components/progress";

  export let organization: string;
  export let quotas: V1Quotas;

  $: projects = createAdminServiceListProjectsForOrganization(organization);

  $: singleProjectLimit = quotas.projects === "1";
  $: dataSize = singleProjectLimit
    ? ""
    : quotas.storageLimitBytesPerDeployment
      ? formatDataSizeQuota(quotas.storageLimitBytesPerDeployment)
      : "";
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">Projects</div>
    <div>
      {$projects.data?.projects?.length} of {quotas.projects ?? "Unlimited"}
    </div>
  </div>

  <div class="quota-entry">
    <div class="quota-entry-title">Data Size {dataSize}</div>
    <div>
      {#if singleProjectLimit}
        <Progress
          value={0}
          max={Number(quotas.storageLimitBytesPerDeployment)}
        />
      {:else}
        <Button type="link" href="/#todo"></Button>
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .quotas {
    @apply flex flex-row items-center;
  }

  .quota-entry {
    @apply flex flex-col;
  }

  .quota-entry-title {
    @apply font-semibold text-[10px] uppercase;
  }
</style>
