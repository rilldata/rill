<script lang="ts">
  import {
    createColumnsProfileData,
    setColumnsProfileStore,
  } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import {
    createQueryServiceTableColumns,
    V1TableColumnsResponse,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";

  export let objectName: string;

  const columnsProfile = createColumnsProfileData();
  setColumnsProfileStore(columnsProfile);

  let profileColumns: CreateQueryResult<V1TableColumnsResponse>;
  $: profileColumns = createQueryServiceTableColumns(
    $runtime?.instanceId,
    objectName,
    {},
    { query: { keepPreviousData: true } }
  );

  $: if ($profileColumns?.data && !$profileColumns?.isFetching) {
    columnsProfile.load($runtime?.instanceId, objectName, $profileColumns);
  }
</script>

<slot />
