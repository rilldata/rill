<script lang="ts">
  import { browser } from "$app/env";
  import LeaderboardContainer from "./_components/LeaderboardContainer.svelte";
  import LeaderboardDisplay from "./_components/LeaderboardDisplay.svelte";
  import LeaderboardHeader from "./_components/LeaderboardHeader.svelte";

  import { getContext, setContext } from "svelte";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  import { createLeaderboardStore } from "./_store";

  let store;
  let persistentTableStore;
  let leaderboardStore;

  if (browser) {
    store = getContext("rill:app:store");
    persistentTableStore = getContext("rill:app:persistent-table-store");
    leaderboardStore = createLeaderboardStore(store.socket);
    setContext("rill:app:leaderboard-store", leaderboardStore);
  }

  /** initialize the leaderboard store */
  $: if (
    !$leaderboardStore?.activeEntityID &&
    $persistentTableStore?.entities?.length
  ) {
    leaderboardStore.setActiveEntityID($persistentTableStore?.entities[0].id);
    leaderboardStore.initializeActiveValues();

    leaderboardStore.socket.emit("getAvailableDimensions", {
      entityType: EntityType.Table,
      entityID: $leaderboardStore.activeEntityID,
    });
    leaderboardStore.socket.emit("getBigNumber", {
      entityType: EntityType.Table,
      entityID: $leaderboardStore.activeEntityID,
      expression: "count(*)",
    });
  }

  /** State for the reference value toggle */
  let whichReferenceValue: string;
  $: stagedReferenceValue =
    whichReferenceValue === "filtered"
      ? $leaderboardStore?.bigNumber
      : $leaderboardStore?.referenceValue;
</script>

<LeaderboardContainer let:columns>
  <LeaderboardHeader bind:whichReferenceValue />
  <LeaderboardDisplay {columns} referenceValue={stagedReferenceValue} />
</LeaderboardContainer>
