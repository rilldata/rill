<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { escapeHtml } from "@rilldata/web-common/lib/i18n";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import type { ReportValues } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import type { Readable } from "svelte/store";
  import type { SuperFormErrors } from "sveltekit-superforms/client";
  import type { ReportFormMetadataProvider } from "@rilldata/web-common/features/scheduled-reports/ReportFormMetadataProvider.svelte.ts";

  let {
    provider,
    data,
    errors,
  }: {
    provider: ReportFormMetadataProvider;
    data: Readable<ReportValues>;
    errors: SuperFormErrors<ReportValues>;
  } = $props();
  let { hasSlackNotifier } = $derived(provider);
</script>

<MultiInput
  id="emailRecipients"
  label={m.report_form_email_recipients()}
  hint={m.report_form_email_hint()}
  bind:values={$data["emailRecipients"]}
  errors={$errors["emailRecipients"]}
  singular="email"
  plural="emails"
  placeholder={m.report_form_email_placeholder()}
/>

{#if hasSlackNotifier}
  <FormSection
    bind:enabled={$data["enableSlackNotification"]}
    showSectionToggle
    title={m.report_form_slack_title()}
    padding=""
  >
    <MultiInput
      id="slackChannels"
      label={m.report_form_channels()}
      hint={m.report_form_slack_channels_hint()}
      bind:values={$data["slackChannels"]}
      errors={$errors["slackChannels"]}
      singular="channel"
      plural="channels"
      placeholder={m.alert_form_slack_placeholder()}
    />
    <MultiInput
      id="slackUsers"
      label={m.report_form_slack_users()}
      hint={m.report_form_slack_users_hint()}
      bind:values={$data["slackUsers"]}
      errors={$errors["slackUsers"]}
      singular="user"
      plural="users"
      placeholder={m.report_form_email_placeholder()}
    />
  </FormSection>
{:else}
  <FormSection title={m.report_form_slack_title()} padding="">
    <svelte:fragment slot="description">
      <span class="text-sm text-fg-secondary">
        {@html m.report_form_slack_not_configured({
          link: `<a href="https://docs.rilldata.com/guides/alerts#configuring-slack-targets" target="_blank">${escapeHtml(m.report_form_docs())}</a>`,
        })}
      </span>
    </svelte:fragment>
  </FormSection>
{/if}
