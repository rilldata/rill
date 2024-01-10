<script lang="ts">
  import OrganizationAccessControls from "./OrganizationAccessControls.svelte";
  import ShareOrganizationCta from "./ShareOrganizationCTA.svelte";

  export let organization: string;
</script>

<div class="flex flex-col gap-y-1">
  <h1 class="text-sky-950 text-4xl font-light leading-10">
    {organization}
  </h1>
  <OrganizationAccessControls {organization}>
    <svelte:fragment slot="manage-org-members">
      <div class="mt-8 mb-12">
        <ShareOrganizationCta {organization} />
      </div>
    </svelte:fragment>
  </OrganizationAccessControls>
  <span class="text-gray-500 text-base font-normal leading-normal">
    Check out your projects below.
  </span>
</div>
