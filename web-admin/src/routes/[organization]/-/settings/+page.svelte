<script lang="ts">
  import { page } from "$app/stores";
  import UpdateOrgItem from "@rilldata/web-admin/features/organizations/settings/UpdateOrgItem.svelte";
  import UserDomainAllowListItem from "@rilldata/web-admin/features/organizations/settings/UserDomainAllowListItem.svelte";

  $: organization = $page.params.organization;
</script>

<div class="flex flex-col gap-y-6">
  <UpdateOrgItem {organization} />
  <UserDomainAllowListItem {organization} />
  <!-- disabling for now since  there are some open questions around billing -->
  <!--  <DeleteOrgItems {organization} />-->
</div>
