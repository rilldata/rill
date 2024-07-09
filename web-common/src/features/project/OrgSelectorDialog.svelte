<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
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

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Select org to deploy to</AlertDialogTitle>
      <Select
        bind:value={selectedOrg}
        id="deploy-target-org"
        label=""
        options={orgs.map((o) => ({ value: o, label: o }))}
      />
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button type="primary" on:click={selectHandler}>Deploy</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
