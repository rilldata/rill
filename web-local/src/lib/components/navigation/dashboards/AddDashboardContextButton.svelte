<script>
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import AddIcon from "@rilldata/web-common/components/icons/Add.svelte";
  import ContextButton from "../../column-profile/ContextButton.svelte";
  import CreateNewDashboardMenu from "./CreateNewDashboardMenu.svelte";
</script>

<WithTogglableFloatingElement
  let:toggleFloatingElement={toggle}
  let:active
  location="right"
  alignment="start"
  distance={24}
>
  <ContextButton
    id="add-dashboard"
    tooltipText="Add a new dashboard"
    on:click={toggle}
    suppressTooltip={active}
    width={24}
    height={24}
    rounded
  >
    <AddIcon />
  </ContextButton>
  <CreateNewDashboardMenu
    slot="floating-element"
    on:escape={toggle}
    on:click-outside={toggle}
    on:bootstrap-dashboard
    on:create-empty-dashboard
  />
</WithTogglableFloatingElement>
