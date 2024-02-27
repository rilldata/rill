<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { getName } from "../entity-management/name-utils";
  import {
    ResourceKind,
    useFilteredResourceNames,
  } from "../entity-management/resource-selectors";
  import CustomDashboardMenuItems from "./CustomDashboardMenuItems.svelte";
  import { createCustomDashboard } from "./createCustomDashboard";

  $: customDashboardNames = useFilteredResourceNames(
    $runtime.instanceId,
    ResourceKind.Dashboard,
  );

  let showCustomDashboards = true;

  async function handleAddCustomDashboard() {
    const newCustomDashboardName = getName(
      "dashboard",
      $customDashboardNames.data ?? [],
    );
    await createCustomDashboard($runtime.instanceId, newCustomDashboardName);
    await goto(`/custom-dashboard/${newCustomDashboardName}`);
  }
</script>

<NavigationHeader
  bind:show={showCustomDashboards}
  toggleText="custom dashboards"
>
  Custom Dashboards
</NavigationHeader>

{#if showCustomDashboards}
  <div
    class="pb-3 max-h-96 overflow-auto"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
  >
    {#if $customDashboardNames?.data}
      {#each $customDashboardNames.data as customDashboardName (customDashboardName)}
        <div
          animate:flip={{ duration: 200 }}
          out:slide|global={{ duration: LIST_SLIDE_DURATION }}
        >
          <NavigationEntry
            name={customDashboardName}
            href={`/custom-dashboard/${customDashboardName}`}
            open={$page.url.pathname ===
              `/custom-dashboard/${customDashboardName}`}
            expandable={false}
          >
            <svelte:fragment slot="menu-items">
              <CustomDashboardMenuItems {customDashboardName} />
            </svelte:fragment>
          </NavigationEntry>
        </div>
      {/each}
    {/if}
    <AddAssetButton
      id="add-custom-dashboard"
      label="Add dashboard"
      bold={false}
      on:click={handleAddCustomDashboard}
    />
  </div>
{/if}
