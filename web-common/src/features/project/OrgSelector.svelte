<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";

  export let open = false;
  export let orgs: string[];
  export let onSelect: (org: string) => void;

  let selectedOrg = "";

  function selectHandler() {
    open = false;
    onSelect(selectedOrg);
  }
</script>

<div class="text-xl">Select an organization</div>
<div class="text-base text-gray-500">
  Choose an organization to deploy this project to. <a
    href="https://docs.rilldata.com/reference/cli/org"
    target="_blank">See docs</a
  >
</div>
<Select
  bind:value={selectedOrg}
  id="deploy-target-org"
  label=""
  placeholder="Select organization"
  options={orgs.map((o) => ({ value: o, label: o }))}
  width={400}
/>
<Button wide type="primary" on:click={selectHandler}>Continue</Button>
