<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceIssueMagicAuthToken } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import CliCommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  $: ({ organization, project, dashboard } = $page.params);

  let token: string;
  let setExpiration = false;

  const {
    dashboardStore,
    selectors: {
      measures: { visibleMeasures },
      dimensions: { visibleDimensions },
    },
  } = getStateManagers();

  const formId = "create-shareable-url-form";

  const initialValues = {
    expiresAt: null,
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
  });

  const issueMagicAuthToken = createAdminServiceIssueMagicAuthToken();

  const { form, enhance, submit, allErrors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        const { token: _token } = await $issueMagicAuthToken.mutateAsync({
          organization,
          project,
          data: {
            metricsView: dashboard,
            metricsViewFilter: $dashboardStore.whereFilter,
            metricsViewFields: [
              ...$visibleMeasures.map((measure) => measure.name),
              ...$visibleDimensions.map((dimension) => dimension.name),
            ],
            ttlMinutes: setExpiration
              ? convertDateToMinutes(values.expiresAt).toString()
              : "0",
          },
        });
        token = _token;
      },
    },
  );

  $: if (setExpiration) {
    // The expiration time should default to 60 days from today
    $form.expiresAt = new Date(Date.now() + 60 * 24 * 60 * 60 * 1000)
      .toISOString()
      .slice(0, 10); // ISO string formatted for input[type="date"]
  } else {
    $form.expiresAt = null;
  }

  $: ({ expiresAt } = $form);
  $: ({ length: allErrorsLength } = $allErrors);

  function convertDateToMinutes(date: string) {
    const now = new Date();
    const future = new Date(date);
    const diff = future.getTime() - now.getTime();
    return Math.floor(diff / 60000);
  }

  $: console.log("$dashboardStore.whereFilter", $dashboardStore.whereFilter);
</script>

{#if !token}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <h3>Create a shareable public link for this view</h3>

    <ul>
      <li>Measures and dimensions will be limited to current visible set.</li>
      <li>Filters will be locked and hidden.</li>
      {#if $dashboardStore.whereFilter}
        <div class="mt-2 px-[19px]">
          <FilterChipsReadOnly
            metricsViewName={dashboard}
            filters={$dashboardStore.whereFilter}
            dimensionThresholdFilters={[]}
            timeRange={undefined}
            comparisonTimeRange={undefined}
          />
        </div>
      {/if}
    </ul>

    <!-- Expiration -->
    <div>
      <div class="has-expiration-container">
        <Switch small id="has-expiration" bind:checked={setExpiration} />
        <Label class="text-xs" for="has-expiration">Set expiration</Label>
      </div>
      {#if setExpiration}
        <div class="expires-at-container">
          <label for="expires-at" class="expires-at-label">
            Access expires
          </label>
          <!-- TODO: use a Rill date picker, once we have one that can select a single day -->
          <input id="expires-at" type="date" bind:value={expiresAt} />
        </div>
      {/if}
    </div>

    <Button type="primary" disabled={$submitting} form={formId} submitForm>
      Create
    </Button>

    {#if allErrorsLength > 0}
      {#each $allErrors as error (error.path)}
        <div class="text-red-500">{error.messages}</div>
      {/each}
    {/if}
  </form>
{:else}
  <!-- A successful form submission will result in a CLI command to display -->
  <CliCommandDisplay
    command={`${window.location.origin}/${organization}/${project}/-/share/${token}`}
  />
{/if}

<style lang="postcss">
  form {
    @apply flex flex-col gap-y-6;
  }

  ul {
    @apply list-disc list-inside;
  }

  .has-expiration-container {
    @apply flex items-center gap-x-2;
  }

  .expires-at-container {
    @apply pl-[30px] mt-2;
    @apply flex items-center gap-x-2;
  }

  .expires-at-label {
    @apply text-slate-500 font-medium;
  }
</style>
