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

  let token: string;

  $: ({ organization, project, dashboard } = $page.params);
  const { dashboardStore } = getStateManagers();

  const formId = "create-shareable-url-form";
  const issueMagicAuthToken = createAdminServiceIssueMagicAuthToken();

  const initialValues = {
    expiresAt: null,
  };

  const validationSchema = object({
    expiresAt: string().nullable(),
  });

  const { form, enhance, submit, allErrors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      async onUpdate({ form }) {
        console.log("submitting form", form);
        const { token: _token } = await $issueMagicAuthToken.mutateAsync({
          organization,
          project,
          data: {
            ttlMinutes: 0, // no expiry
            metricsView: dashboard,
            metricsViewFilter: undefined,
            metricsViewFields: undefined,
          },
        });
        token = _token;
      },
    },
  );

  $: ({ length: allErrorsLength } = $allErrors);

  let showExpiration = false;
</script>

{#if token}
  <CliCommandDisplay
    command={`${window.location.origin}/${organization}/${project}/-/share/${token}`}
  />
{:else}
  <form id={formId} on:submit|preventDefault={submit} use:enhance>
    <h3>Create a shareable public link for this view</h3>

    <ul>
      <li>Measures and dimensions will be limited to current visible set.</li>
      <li>Filters will be locked and hidden.</li>
      {#if $dashboardStore.whereFilter || $dashboardStore.dimensionThresholdFilters.length > 0}
        <div class="mt-2 px-[19px]">
          <FilterChipsReadOnly
            metricsViewName={dashboard}
            filters={$dashboardStore.whereFilter}
            dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
            timeRange={undefined}
            comparisonTimeRange={undefined}
          />
        </div>
      {/if}
    </ul>

    <!-- Expiration -->
    <div>
      <div class="flex items-center gap-x-2">
        <Switch small id="expiration" bind:checked={showExpiration} />
        <div class="flex flex-col">
          <Label class="text-xs" for="expiration">Set expiration</Label>
        </div>
      </div>
      {#if showExpiration}
        <div class="pl-[30px] mt-2 flex items-center gap-x-2">
          <div class="text-slate-500 font-medium">Access expires</div>
          <input type="date" bind:value={$form.expiresAt} />
          <!-- TODO: use a Rill date picker -->
          <!-- <IconButton on:click={() => (showDatePicker = true)}>
            <EditIcon className="text-primary-500" />
          </IconButton> -->
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
{/if}

<!-- Result -->

<style lang="postcss">
  form {
    @apply flex flex-col gap-y-6;
  }

  ul {
    @apply list-disc list-inside;
  }
</style>
