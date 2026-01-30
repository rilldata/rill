<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceIssueMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
    type AdminServiceIssueMagicAuthTokenBody,
  } from "@rilldata/web-admin/client";
  import { isCanvasDashboardPage } from "@rilldata/web-admin/features/navigation/nav-utils";
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
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Pencil } from "lucide-svelte";
  import { DateTime } from "luxon";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import CanvasFiltersSection from "./CanvasFiltersSection.svelte";
  import ExploreFiltersSection from "./ExploreFiltersSection.svelte";
  import { convertDateToMinutes } from "./form-utils";

  const queryClient = useQueryClient();

  $: ({ organization, project, dashboard } = $page.params);
  $: ({ instanceId } = $runtime);
  $: isCanvas = isCanvasDashboardPage($page);

  $: isTitleEmpty = $form.title.trim() === "";

  let url: string | null = null;
  let setExpiration = false;
  let apiError: string;
  let popoverOpen = false;
  let copied = false;

  // These will be set by the child components via callbacks
  let hasSomeFilter = false;
  let dashboardDataProvider:
    | (() => Partial<AdminServiceIssueMagicAuthTokenBody>)
    | null = null;

  function handleFilterStateChange(hasFilters: boolean) {
    hasSomeFilter = hasFilters;
  }

  function handleProvideFilters(
    provider: () => Partial<AdminServiceIssueMagicAuthTokenBody>,
  ) {
    dashboardDataProvider = provider;
  }

  const formId = "create-public-url-form";

  const initialValues = {
    expiresAt: null,
    title: "",
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
    title: string().required("Title is required"),
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
          if (!dashboardDataProvider) {
            throw new Error("Dashboard data provider not initialized");
          }

          const dashboardData = dashboardDataProvider();

          const { url: _url } = await $issueMagicAuthToken.mutateAsync({
            org: organization,
            project,
            data: {
              ...dashboardData,
              resourceName: dashboard,
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

  function onCopy() {
    copyToClipboard(url, "URL copied to clipboard", false);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }

  $: if (setExpiration && $form.expiresAt === null) {
    // When `setExpiration` is toggled, initialize the expiration time to 60 days from today
    $form.expiresAt = DateTime.now().plus({ days: 60 }).toISO();
    popoverOpen = true;
  } else if (!setExpiration) {
    $form.expiresAt = null;
    popoverOpen = false;
  }

  $: ({ length: allErrorsLength } = $allErrors);

  $: maxExpirationDate = DateTime.now().plus({ years: 1 }).startOf("day");
</script>

{#if !url}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <h3 class="text-xs text-fg-primary font-normal">
      Create a shareable public URL for this view.
    </h3>

    <div class="flex flex-col gap-y-1 mt-4">
      <Input
        id="name-input"
        bind:value={$form.title}
        placeholder="Label this URL"
      />
    </div>

    <div class="mt-4" class:mb-4={!hasSomeFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration">Set expiration</Label>
      </div>
      {#if setExpiration}
        <div class="flex items-center gap-x-1 pl-[30px]">
          <label for="expires-at" class="text-fg-secondary font-medium">
            Access expires {new Date($form.expiresAt).toLocaleDateString(
              "en-US",
              { year: "numeric", month: "short", day: "numeric" },
            )}
          </label>
          <Popover bind:open={popoverOpen}>
            <PopoverTrigger>
              <IconButton ariaLabel="Edit expiration date">
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

    <!-- TODO: revisit when time range lock is implemented -->
    <!-- <div class="mt-4" class:mb-4={!hasWhereFilter}>
      <div class="flex items-center gap-x-2">
        <Switch small id="lock-time-range" bind:checked={lockTimeRange} />

        <div class="flex flex-row items-center gap-x-1">
          <Label class="text-xs" for="lock-time-range">Lock time range</Label>
          <Tooltip location="right" alignment="middle" distance={8}>
            <div class="text-fg-secondary">
              <InfoCircle size="12px" />
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              Only data within this range will be visible
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
      {#if lockTimeRange}
        <div class="w-full pl-[30px]">
          <label for="lock-time-range" class="text-fg-secondary font-medium">
            {#if interval.isValid}
              <RangeDisplay {interval} grain={activeTimeGrain} {abbreviation} />
            {/if}
          </label>
        </div>
      {/if}
    </div> -->

    {#if isCanvas && instanceId}
      <CanvasFiltersSection
        {dashboard}
        {instanceId}
        onFilterStateChange={handleFilterStateChange}
        onProvideFilters={handleProvideFilters}
      />
    {:else}
      <ExploreFiltersSection
        onFilterStateChange={handleFilterStateChange}
        onProvideFilters={handleProvideFilters}
      />
    {/if}

    <Button
      type="primary"
      disabled={$submitting || isTitleEmpty}
      form={formId}
      submitForm
    >
      Create
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
    <h3>Success! A public URL has been created.</h3>
    <Button
      type="secondary"
      onClick={onCopy}
      dataAttributes={{ "data-public-url": url }}
    >
      {#if copied}
        <Check size="16px" />
        Copied URL
      {:else}
        Copy Public URL
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
