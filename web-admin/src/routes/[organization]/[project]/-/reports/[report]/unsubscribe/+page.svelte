<script lang="ts">
  import { page } from "$app/stores";
  import { type RpcStatus } from "@rilldata/web-admin/client";
  import { createAdminServiceUnsubscribeReportUsingToken } from "@rilldata/web-admin/features/scheduled-reports/unsubscribe-report-using-token";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import type { AxiosError } from "axios";
  import { onMount } from "svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: report = $page.params.report;
  $: token = $page.url.searchParams.get("token");

  // using this instead of reportUnsubscriber to avoid a flicker before reportUnsubscriber is triggered
  let loading = true;

  const reportUnsubscriber = createAdminServiceUnsubscribeReportUsingToken();

  $: error =
    ($reportUnsubscriber.error as unknown as AxiosError<RpcStatus>)?.response
      ?.data?.message ?? $reportUnsubscriber.error?.message;

  async function unsubscribe() {
    await $reportUnsubscriber.mutateAsync({
      organization,
      project,
      name: report,
      token,
      data: {},
    });
    loading = false;
  }

  onMount(() => unsubscribe());
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="flex flex-col gap-y-2">
      {#if error}
        <h2 class="text-lg font-semibold">Failed to unsubscribe.</h2>
        <CtaMessage>
          {error}
        </CtaMessage>
      {:else if loading}
        <h2 class="text-lg font-semibold">Unsubscribing...</h2>
      {:else}
        <h2 class="text-lg font-semibold">Unsubscribed from report.</h2>
      {/if}
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
