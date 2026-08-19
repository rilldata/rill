<script lang="ts">
  import { page } from "$app/state";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil } from "lucide-svelte";
  import { DateTime } from "luxon";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import {
    convertDateToMinutes,
    createFieldsAndStateForKind,
  } from "./form-utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { getLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
  import {
    CanvasDashboardConfigProvider,
    ExploreDashboardConfigProvider,
  } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();

  let { organization, project, dashboard } = $derived(page.params);

  let url = $state<string | null>(null);
  let setExpiration = $state(false);
  let apiError: string = $state("");
  let popoverOpen = $state(false);
  let copied = $state(false);

  // Get expression filter manager for the resource.
  const dashboardResource = getDashboardResourceFromPage(page);
  const isExplore = dashboardResource?.kind === ResourceKind.Explore;

  const dashboardConfigProvider = isExplore
    ? new ExploreDashboardConfigProvider(runtimeClient, dashboardResource?.name)
    : new CanvasDashboardConfigProvider(runtimeClient, dashboardResource?.name);
  const expressionFilterManager = new ExpressionFilterManager(
    dashboardConfigProvider.metricsViewsProvider,
    dashboardConfigProvider.yamlConfigProvider,
  );
  const { setUrlParams } = expressionFilterManager;

  let curUrlParams = $derived(page.url.searchParams);
  $effect(() => setUrlParams(curUrlParams));

  const exprByMetricsView = $derived(expressionFilterManager.exprByMetricsView);
  const hasSomeFilter = $derived(Object.keys(exprByMetricsView).length > 0);

  const sanitisedFilterState = createFieldsAndStateForKind(
    dashboardResource?.kind,
    expressionFilterManager,
  );
  let { fields, sanitizedState, queryTimeStart, queryTimeEnd } = $derived(
    $sanitisedFilterState,
  );
  $effect(() => console.log($sanitisedFilterState));

  const formId = "create-public-url-form";

  const initialValues = {
    expiresAt: null,
    title: "",
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
    title: string().required(m.public_url_title_required()),
  });

  const issueMagicAuthToken = createAdminServiceIssueMagicAuthToken();

  const { form, enhance, submit, allErrors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          const { url: _url } = await $issueMagicAuthToken.mutateAsync({
            org: organization,
            project,
            data: {
              resourceType: dashboardResource?.kind,
              resourceName: dashboard,
              metricsViewFilters: exprByMetricsView,
              fields,
              state: sanitizedState,
              ttlMinutes: setExpiration
                ? convertDateToMinutes(values.expiresAt).toString()
                : undefined,
              displayName: values.title,
            },
          });

          url = _url;

          void queryClient.invalidateQueries({
            queryKey: getAdminServiceListMagicAuthTokensQueryKey(
              organization,
              project,
            ),
          });
        } catch (error) {
          const typedError = error as HTTPError;
          apiError = typedError.response?.data?.message ?? typedError.message;
        }
      },
      invalidateAll: false,
    },
  );

  let isTitleEmpty = $derived($form.title.trim() === "");

  $effect(() => {
    if (setExpiration && $form.expiresAt === null) {
      // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
      $form.expiresAt = DateTime.now().plus({ days: 60 }).toISO();
      popoverOpen = true;
    } else if (!setExpiration) {
      $form.expiresAt = null;
      popoverOpen = false;
    }
  });

  let { length: allErrorsLength } = $derived($allErrors);
  let maxExpirationDate = $derived(
    DateTime.now().plus({ years: 1 }).startOf("day"),
  );

  function onCopy() {
    copyToClipboard(url, "URL copied to clipboard", false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

{#if !url}
  <form
    id={formId}
    onsubmit={(e) => {
      e.preventDefault();
      submit(e);
    }}
    use:enhance
  >
    <h3 class="text-xs text-fg-primary font-normal">
      {m.public_url_create_heading()}
    </h3>

    <div class="flex flex-col gap-y-1 mt-4">
      <Input
        id="name-input"
        bind:value={$form.title}
        placeholder={m.public_url_label_placeholder()}
      />
    </div>

    <div class="mt-4" class:mb-4={!hasSomeFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration"
          >{m.public_url_set_expiration()}</Label
        >
      </div>
      {#if setExpiration}
        <div class="flex items-center gap-x-1 pl-[30px]">
          <label for="expires-at" class="text-fg-secondary font-medium">
            {m.public_url_access_expires({
              date: new Date($form.expiresAt).toLocaleDateString(getLocale(), {
                year: "numeric",
                month: "short",
                day: "numeric",
              }),
            })}
          </label>
          <Popover bind:open={popoverOpen}>
            <PopoverTrigger>
              <IconButton ariaLabel={m.public_url_edit_expiration()}>
                <Pencil size="14px" class="text-primary-600" />
              </IconButton>
            </PopoverTrigger>
            <PopoverContent align="end" class="p-0" strategy="fixed">
              <Calendar
                selection={DateTime.fromISO($form.expiresAt)}
                singleDaySelection
                maxDate={maxExpirationDate}
                firstVisibleMonth={DateTime.fromISO($form.expiresAt)}
                onSelectDay={(date) => {
                  $form.expiresAt = date.toISO();
                  popoverOpen = false;
                }}
              />
            </PopoverContent>
          </Popover>
        </div>
      {/if}
    </div>

    {#if hasSomeFilter}
      <hr class="mt-4 mb-4" />

      <div class="flex flex-col gap-y-1">
        <p class="text-xs text-fg-primary font-normal">
          {m.public_url_filters_locked_hidden()}
        </p>
        <div class="flex flex-col gap-2 my-2">
          <ReadonlyExpressionFilters
            {expressionFilterManager}
            {queryTimeStart}
            {queryTimeEnd}
          />
        </div>
      </div>

      {#if isExplore}
        <p class="text-xs text-fg-primary font-normal mt-4 mb-4">
          {m.public_url_measures_dimensions_limited()}
        </p>
      {/if}
    {/if}

    <Button
      type="primary"
      disabled={$submitting || isTitleEmpty}
      form={formId}
      submitForm
    >
      {m.public_url_create_button()}
    </Button>

    {#if allErrorsLength > 0}
      {#each $allErrors as error (error.path)}
        <div class="text-red-500 mt-1">{error.messages}</div>
      {/each}
    {:else if apiError}
      <div class="text-red-500 mt-1">{apiError}</div>
    {/if}
  </form>
{:else}
  <div class="flex flex-col gap-y-4">
    <h3>{m.public_url_success_message()}</h3>
    <Button
      type="secondary"
      onClick={onCopy}
      dataAttributes={{ "data-public-url": url }}
    >
      {#if copied}
        <Check size="16px" />
        {m.public_url_copied_button()}
      {:else}
        {m.public_url_copy_button()}
      {/if}
    </Button>
  </div>
{/if}

<style lang="postcss">
  form {
    @apply flex flex-col;
  }

  h3 {
    @apply font-semibold;
  }
</style>
