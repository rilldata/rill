<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import AlertTriangle from "@rilldata/web-common/components/icons/AlertTriangle.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils";
  import { getSnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
  import ScheduleForm from "@rilldata/web-common/features/scheduled-reports/ScheduleForm.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import ExportFileOptions from "@rilldata/web-common/features/exports/ExportFileOptions.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { SuperForm } from "sveltekit-superforms/client";

  export let superFormInstance: SuperForm<AlertFormValues>;
  export let exploreName: string;

  const runtimeClient = useRuntimeClient();

  $: ({ form, errors } = superFormInstance);

  $: hasSlackNotifier = getHasSlackConnection(runtimeClient);

  // Time zone options from the dashboard's spec
  $: exploreSpecQuery = useExploreValidSpec(runtimeClient, exploreName);
  $: availableTimeZones = $exploreSpecQuery.data?.explore?.timeZones;
</script>

<div class="flex flex-col gap-y-3">
  <FormSection title={m.alert_form_name_title()}>
    <Input
      alwaysShowError
      errors={$errors["name"]}
      id="name"
      title={m.alert_form_name_title()}
      placeholder={m.alert_form_name_placeholder()}
      bind:value={$form["name"]}
    />
  </FormSection>

  <FormSection title={m.alert_form_trigger()}>
    <div class="grid grid-cols-2">
      <Button
        onClick={() => ($form["refreshWhenDataRefreshes"] = true)}
        selected={$form["refreshWhenDataRefreshes"]}
      >
        {m.alert_form_trigger_data_refresh()}
      </Button>
      <Button
        onClick={() => ($form["refreshWhenDataRefreshes"] = false)}
        selected={!$form["refreshWhenDataRefreshes"]}
      >
        {m.alert_form_trigger_set_schedule()}
      </Button>
    </div>
    {#if !$form["refreshWhenDataRefreshes"]}
      <ScheduleForm data={form} {availableTimeZones} />
    {/if}
  </FormSection>

  <FormSection
    description={m.alert_form_snooze_desc()}
    title={m.alert_form_snooze_title()}
  >
    <Select
      bind:value={$form["snooze"]}
      id="snooze"
      label=""
      options={getSnoozeOptions()}
    />
  </FormSection>
  {#if $hasSlackNotifier.data}
    <FormSection
      bind:enabled={$form["enableSlackNotification"]}
      showSectionToggle
      title={m.alert_form_slack_title()}
    >
      <MultiInput
        id="slackChannels"
        placeholder={m.alert_form_slack_placeholder()}
        description={m.alert_form_slack_channels_desc()}
        contentClassName="relative"
        bind:values={$form.slackChannels}
        errors={$errors.slackChannels}
        singular="channel"
        plural="channels"
        preventFocus={true}
      />
      <MultiInput
        id="slackUsers"
        placeholder={m.alert_form_email_placeholder()}
        description={m.alert_form_slack_users_desc()}
        contentClassName="relative"
        bind:values={$form.slackUsers}
        errors={$errors.slackUsers}
        singular="user"
        plural="users"
        preventFocus={true}
      />
    </FormSection>
  {:else}
    <FormSection title={m.alert_form_slack_title()}>
      <svelte:fragment slot="description">
        <span class="text-sm text-fg-secondary">
          {@html m.alert_form_slack_not_configured({
            docsUrl:
              "https://docs.rilldata.com/guide/alerts#configuring-slack-targets",
          })}
        </span>
      </svelte:fragment>
    </FormSection>
  {/if}
  <FormSection
    bind:enabled={$form["enableEmailNotification"]}
    description={m.alert_form_email_desc()}
    showSectionToggle
    title={m.alert_form_email_title()}
  >
    <MultiInput
      id="slackUsers"
      placeholder={m.alert_form_email_placeholder()}
      contentClassName="relative"
      bind:values={$form.emailRecipients}
      errors={$errors.emailRecipients}
      singular="email"
      plural="emails"
      preventFocus={true}
    />
    <Checkbox
      bind:checked={$form["exportAttach"]}
      id="exportAttach"
      onCheckedChange={(checked) => {
        $form["exportAttach"] = Boolean(checked);
      }}
      inverse
      label={m.alert_form_attach_rows()}
      hint={m.alert_form_attach_rows_hint()}
    />
    {#if $form["exportAttach"]}
      <div
        class="flex items-start gap-x-2 rounded-sm border border-yellow-200 bg-yellow-50 px-3 py-2 text-xs text-yellow-800 dark:border-yellow-900/40 dark:bg-yellow-900/20 dark:text-yellow-300"
      >
        <AlertTriangle size="14px" className="mt-px flex-none" />
        <span>{m.alert_form_attach_rows_warning()}</span>
      </div>
      <ExportFileOptions
        bind:format={$form["exportFormat"]}
        bind:limit={$form["exportLimit"]}
        bind:includeHeader={$form["exportIncludeHeader"]}
        limitErrors={$errors["exportLimit"]}
      />
    {/if}
  </FormSection>
</div>
