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
  Choose an organization to deploy this project to. See docs
</div>
<Select
  bind:value={selectedOrg}
  id="deploy-target-org"
  label=""
  placeholder="Select organization"
  options={orgs.map((o) => ({ value: o, label: o }))}
  className="w-[400px]"
/>
<Button wide type="primary" on:click={selectHandler}>Continue</Button>
