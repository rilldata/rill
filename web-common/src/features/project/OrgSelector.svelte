<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import {
    SelectSeparator,
    SelectItem,
  } from "@rilldata/web-common/components/select";

  export let open = false;
  export let orgs: string[];
  export let onSelect: (org: string) => void;
  export let onNewOrg: () => void;

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
<!-- w-[400px] Needed for tailwind to compile for this -->
<Select
  bind:value={selectedOrg}
  id="deploy-target-org"
  label=""
  placeholder="Select organization"
  options={orgs.map((o) => ({ value: o, label: o }))}
  width={400}
  sameWidth
>
  <div slot="extra-dropdown-content">
    <SelectSeparator />
    <SelectItem
      value="__rill_new_org"
      on:click={onNewOrg}
      class="text-[12px] gap-x-2 items-start"
    >
      + Create organization
    </SelectItem>
  </div>
</Select>
<Button wide type="primary" on:click={selectHandler}>Continue</Button>
