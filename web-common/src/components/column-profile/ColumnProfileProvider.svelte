<script lang="ts">
  import {
    createColumnsProfileData,
    setColumnsProfileStore,
  } from "@rilldata/web-common/components/column-profile/columns-profile-data";
  import { getTableReferences } from "@rilldata/web-common/features/models/utils/get-table-references";
  import { getMatchingReferencesAndEntries } from "@rilldata/web-common/features/models/workspace/inspector/utils";
  import {
    createQueryServiceTableColumns,
    createRuntimeServiceListCatalogEntries,
    V1TableColumnsResponse,
  } from "@rilldata/web-common/runtime-client";
  import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client/index.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";

  export let objectName: string;
  export let sql = "";

  const columnsProfile = createColumnsProfileData();
  setColumnsProfileStore(columnsProfile);

  let profileColumns: CreateQueryResult<V1TableColumnsResponse>;
  $: profileColumns = createQueryServiceTableColumns(
    $runtime?.instanceId,
    objectName,
    {},
    { query: { keepPreviousData: true } }
  );

  $: getAllSources = createRuntimeServiceListCatalogEntries(
    $runtime?.instanceId,
    {
      type: "OBJECT_TYPE_SOURCE",
    }
  );
  $: getAllModels = createRuntimeServiceListCatalogEntries(
    $runtime?.instanceId,
    {
      type: "OBJECT_TYPE_MODEL",
    }
  );
  let referencedThings: Array<V1CatalogEntry>;
  // for each reference, match to an existing model or source,
  $: if (objectName && sql) {
    referencedThings = getMatchingReferencesAndEntries(
      objectName,
      getTableReferences(sql),
      [
        ...($getAllSources?.data?.entries || []),
        ...($getAllModels?.data?.entries || []),
      ]
    ).map(([entity]) => entity) as Array<V1CatalogEntry>;
  }

  $: if ($profileColumns?.data && !$profileColumns?.isFetching) {
    columnsProfile.load(
      $runtime?.instanceId,
      objectName,
      $profileColumns,
      referencedThings
    );
  }
</script>

<slot />
