<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ConnectYourDataWidget
  startConnectorSelection={(name) =>
    void goto(
      `/${organization}/${project}/-/welcome/add-data?schema=${name ?? ""}`,
    )}
  onWelcomeScreen
/>
