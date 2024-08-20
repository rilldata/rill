<script lang="ts">
  import { page } from "$app/stores";
  import DeleteOrgItems from "@rilldata/web-admin/features/organizations/settings/DeleteOrgItems.svelte";
  import UpdateOrgItem from "@rilldata/web-admin/features/organizations/settings/UpdateOrgItem.svelte";

  $: organization = $page.params.organization;
</script>

<div class="flex flex-col gap-y-6">
  <UpdateOrgItem {organization} />
  <DeleteOrgItems {organization} />
</div>
