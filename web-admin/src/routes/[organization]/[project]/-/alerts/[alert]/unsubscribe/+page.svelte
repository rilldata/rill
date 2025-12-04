<script lang="ts">
  import { page } from "$app/stores";
  import {
    type AdminServiceUnsubscribeAlertBodyBody,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { createAdminServiceUnsubscribeAlertUsingToken } from "@rilldata/web-admin/features/alerts/unsubscribe-alert-using-token.ts";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import type { AxiosError } from "axios";
  import { onMount } from "svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alert = $page.params.alert;
  $: token = $page.url.searchParams.get("token");
  $: email = $page.url.searchParams.get("email");
  $: slackUser = $page.url.searchParams.get("slack_user");

  // using this instead of alertUnsubscriber to avoid a flicker before alertUnsubscriber is triggered
  let loading = true;

  const alertUnsubscriber = createAdminServiceUnsubscribeAlertUsingToken();

  $: error =
    ($alertUnsubscriber.error as unknown as AxiosError<RpcStatus>)?.response
      ?.data?.message ?? $alertUnsubscriber.error?.message;

  async function unsubscribe() {
    const data: AdminServiceUnsubscribeAlertBodyBody = {};
    if (email) data.email = email;
    if (slackUser) data.slackUser = slackUser;

    await $alertUnsubscriber.mutateAsync({
      organization,
      project,
      name: alert,
      data,
      token,
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
        <h2 class="text-lg font-semibold">Unsubscribed from alert.</h2>
      {/if}
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
