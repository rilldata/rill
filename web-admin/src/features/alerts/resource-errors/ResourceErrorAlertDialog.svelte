<script lang="ts">
  import {
    createAdminServiceCreateAlert,
    createAdminServiceEditAlert,
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceListProjectMemberUsers,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { getHasSlackConnection } from "@rilldata/web-common/features/alerts/delivery-tab/notifiers-utils";
  import { getSnoozeOptions } from "@rilldata/web-common/features/alerts/delivery-tab/snooze";
  import { prettyResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    type V1AlertSpec,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { untrack } from "svelte";
  import { ChevronDown } from "lucide-svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import {
    displayKinds,
    pluralizeKind,
  } from "@rilldata/web-admin/features/projects/status/overview/overview-utils";
  import {
    DEFAULT_CHECK_CRON,
    getAlertOptionsFromFormValues,
    getFormValuesFromAlertSpec,
    resourceKey,
    type ResourceErrorAlertFormValues,
  } from "./utils";

  let {
    org,
    project,
    open = $bindable(false),
    alertName = undefined,
    alertSpec = undefined,
    initialResource = undefined,
  }: {
    org: string;
    project: string;
    open?: boolean;
    // Set when editing an existing alert.
    alertName?: string;
    alertSpec?: V1AlertSpec;
    // Pre-selected resource as "kind/name" when creating.
    initialResource?: string;
  } = $props();

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();

  let isEdit = $derived(!!alertName);
  const createAlert = createAdminServiceCreateAlert();
  const editAlert = createAdminServiceEditAlert();

  const currentUser = createAdminServiceGetCurrentUser();
  let projectAdmins = $derived(
    createAdminServiceListProjectMemberUsers(org, project, {
      role: "admin",
      pageSize: 1000,
    }),
  );
  let orgAdmins = $derived(
    createAdminServiceListOrganizationMemberUsers(org, {
      role: "admin",
      pageSize: 1000,
    }),
  );
  const resources = createRuntimeServiceListResources(runtimeClient, {});
  const hasSlackNotifier = getHasSlackConnection(runtimeClient);

  const checkOptions = [
    { value: "*/5 * * * *", label: m.alert_errors_check_5m() },
    { value: "*/15 * * * *", label: m.alert_errors_check_15m() },
    { value: "0 * * * *", label: m.alert_errors_check_hourly() },
  ];

  // The form is initialized once per mount.
  const initialValues: ResourceErrorAlertFormValues = untrack(() =>
    alertSpec
      ? getFormValuesFromAlertSpec(alertSpec)
      : {
          name: initialResource
            ? m.alert_errors_default_name_resource({
                name: initialResource.split("/")[1],
              })
            : m.alert_errors_default_name(),
          resources: initialResource ? [initialResource] : [],
          refreshCron: DEFAULT_CHECK_CRON,
          emailRecipients: [],
          slackChannels: [],
          slackUsers: [],
          snooze: getSnoozeOptions()[0].value,
        },
  );

  const schema = yup(
    object({
      name: string().trim().required(m.alert_errors_name_required()),
      resources: array(string().defined()).defined(),
      refreshCron: string().required(),
      emailRecipients: array(string().defined()).defined(),
      slackChannels: array(string().defined()).defined(),
      slackUsers: array(
        string().defined().email(m.alert_errors_invalid_email()),
      ).defined(),
      snooze: string().defined(),
    }),
  );

  let scopeSelected = $state(initialValues.resources.length > 0);
  let submitError = $state("");

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      resetForm: false,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values: ResourceErrorAlertFormValues = {
          ...form.data,
          resources: scopeSelected ? form.data.resources : [],
        };
        if (
          !values.emailRecipients.length &&
          !values.slackChannels.some(Boolean) &&
          !values.slackUsers.some(Boolean)
        ) {
          submitError = m.alert_errors_recipient_required();
          return;
        }
        if (scopeSelected && !values.resources.length) {
          submitError = m.alert_errors_resource_required();
          return;
        }

        submitError = "";
        try {
          const options = getAlertOptionsFromFormValues(values);
          if (alertName) {
            await $editAlert.mutateAsync({
              org,
              project,
              name: alertName,
              data: { options },
            });
          } else {
            await $createAlert.mutateAsync({ org, project, data: { options } });
          }
        } catch (e) {
          const err = e as { response?: { data?: { message?: string } } };
          submitError = err.response?.data?.message ?? String(e);
          return;
        }

        await queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            runtimeClient.instanceId,
          ),
        });
        open = false;
        eventBus.emit("notification", {
          message: isEdit ? m.alert_errors_edited() : m.alert_errors_created(),
          type: "success",
        });
      },
    },
  );

  // Pre-select the current user as a recipient when creating.
  let didDefaultRecipient = false;
  $effect(() => {
    const email = $currentUser.data?.user?.email;
    if (isEdit || didDefaultRecipient || !email) return;
    didDefaultRecipient = true;
    $form.emailRecipients = [email];
  });

  let adminItems = $derived.by(() => {
    const byEmail = new Map<string, string>();
    for (const u of [
      ...($projectAdmins.data?.members ?? []),
      ...($orgAdmins.data?.members ?? []),
    ]) {
      if (!u.userEmail) continue;
      byEmail.set(
        u.userEmail,
        u.userName ? `${u.userName} (${u.userEmail})` : u.userEmail,
      );
    }
    // Keep recipients that aren't listed (e.g. admins through a user group).
    for (const email of $form.emailRecipients) {
      if (!byEmail.has(email)) byEmail.set(email, email);
    }
    return [...byEmail]
      .map(([name, label]) => ({ name, label }))
      .sort((a, b) => a.label.localeCompare(b.label));
  });

  let resourceGroups = $derived.by(() =>
    displayKinds
      .map((kind) => ({
        name: kind,
        label: pluralizeKind(prettyResourceKind(kind), 2),
        items: ($resources.data?.resources ?? [])
          .filter((r) => r.meta?.name?.kind === kind)
          .map((r) => {
            const name = r.meta?.name?.name ?? "";
            return { name: resourceKey(kind, name), label: name };
          })
          .sort((a, b) => a.label.localeCompare(b.label)),
      }))
      .filter((g) => g.items.length > 0),
  );

  let checkSelectOptions = $derived(
    checkOptions.some((o) => o.value === $form.refreshCron)
      ? checkOptions
      : [
          ...checkOptions,
          { value: $form.refreshCron, label: $form.refreshCron },
        ],
  );

  function toggle(list: string[], value: string): string[] {
    return list.includes(value)
      ? list.filter((v) => v !== value)
      : [...list, value];
  }

  function resourceLabel(key: string): string {
    const [kind, name] = key.split("/");
    return `${prettyResourceKind(kind)}: ${name}`;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>
        {isEdit ? m.alert_errors_edit_title() : m.alert_errors_create_title()}
      </Dialog.Title>
      <Dialog.Description>{m.alert_errors_description()}</Dialog.Description>
    </Dialog.Header>

    <form
      onsubmit={(e) => {
        e.preventDefault();
        submit(e);
      }}
      use:enhance
      class="flex flex-col gap-y-3 pt-2"
    >
      <Input
        bind:value={$form.name}
        id="name"
        label={m.alert_errors_name_label()}
        errors={$errors.name}
        alwaysShowError
      />

      <FormSection title={m.alert_errors_scope_title()}>
        <div class="grid grid-cols-2">
          <Button
            onClick={() => (scopeSelected = false)}
            selected={!scopeSelected}
          >
            {m.alert_errors_scope_all()}
          </Button>
          <Button
            onClick={() => (scopeSelected = true)}
            selected={scopeSelected}
          >
            {m.alert_errors_scope_selected()}
          </Button>
        </div>
        {#if scopeSelected}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <Button {...props} type="secondary" wide>
                  <span class="truncate">
                    {$form.resources.length
                      ? m.alert_errors_resources_selected({
                          count: $form.resources.length,
                        })
                      : m.alert_errors_resources_placeholder()}
                  </span>
                  <ChevronDown size="14px" />
                </Button>
              {/snippet}
            </DropdownMenu.Trigger>
            <SearchableMenuContent
              allowMultiSelect
              allowSelectAll={false}
              selectableGroups={resourceGroups}
              selectedItems={resourceGroups.map((g) =>
                $form.resources.filter((r) => r.startsWith(`${g.name}/`)),
              )}
              onSelect={(key) =>
                ($form.resources = toggle($form.resources, key))}
            />
          </DropdownMenu.Root>
          {#if $form.resources.length}
            <ul class="text-xs text-fg-secondary flex flex-col gap-y-0.5">
              {#each $form.resources as key (key)}
                <li>{resourceLabel(key)}</li>
              {/each}
            </ul>
          {/if}
        {/if}
      </FormSection>

      <Select
        bind:value={$form.refreshCron}
        id="refreshCron"
        label={m.alert_errors_check_label()}
        options={checkSelectOptions}
        sameWidth
      />

      <FormSection
        title={m.alert_errors_email_title()}
        description={m.alert_errors_email_desc()}
      >
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button {...props} type="secondary" wide>
                <span class="truncate">
                  {$form.emailRecipients.length
                    ? $form.emailRecipients.join(", ")
                    : m.alert_errors_email_placeholder()}
                </span>
                <ChevronDown size="14px" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <SearchableMenuContent
            allowMultiSelect
            allowSelectAll={false}
            selectableGroups={[{ name: "admins", items: adminItems }]}
            selectedItems={[$form.emailRecipients]}
            onSelect={(email) =>
              ($form.emailRecipients = toggle($form.emailRecipients, email))}
          />
        </DropdownMenu.Root>
      </FormSection>

      {#if $hasSlackNotifier.data}
        <FormSection title={m.alert_form_slack_title()}>
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
      {/if}

      <Select
        bind:value={$form.snooze}
        id="snooze"
        label={m.alert_form_snooze_title()}
        options={getSnoozeOptions()}
        sameWidth
      />

      {#if submitError}
        <p class="text-destructive text-sm">{submitError}</p>
      {/if}
    </form>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (open = false)}>
        {m.common_cancel()}
      </Button>
      <Button
        type="primary"
        onClick={submit}
        loading={$submitting}
        disabled={$submitting}
      >
        {isEdit ? m.alert_errors_save() : m.alert_errors_create()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
