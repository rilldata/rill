<script lang="ts">
  import { createEventDispatcher, getContext } from "svelte";
  import VirtualizedGrid from "./VirtualizedGrid.svelte";
  import Leaderboard from "./Leaderboard.svelte";
  import { browser } from "$app/env";
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { ExploreStore } from "../_store";
  export let columns: number;
  export let referenceValue: number;

  const leaderboardStore: ExploreStore = getContext(
    "rill:app:leaderboard-store"
  );

  const dispatch = createEventDispatcher();
  let leaderboardExpanded;

  // prepare the activeFilters to be sent to the server
  function prune(actives) {
    return Object.keys(actives)
      .filter((key) => {
        return actives[key].length;
      })
      .reduce((acc, v) => {
        acc[v] = actives[v].map((value) => [value, "include"]);
        return acc;
      }, {});
  }
</script>

<div
  style:height="calc(100vh - var(--header, 130px) - 4rem)"
  class="
      border-t border-gray-200 overflow-auto
      "
>
  {#if $leaderboardStore}
    {#key $leaderboardStore?.activeEntityID}
      <VirtualizedGrid
        {columns}
        height="100%"
        items={$leaderboardStore.leaderboards}
        let:item
      >
        <!-- the single virtual element -->
        <div style:width="315px">
          <Leaderboard
            seeMore={leaderboardExpanded === item.displayName}
            on:expand={() => {
              if (leaderboardExpanded === item.displayName) {
                leaderboardExpanded = undefined;
              } else {
                leaderboardExpanded = item.displayName;
              }
            }}
            on:select-item={(event) => {
              dispatch("select-item", {
                fieldName: event.detail,
                dimensionName: item.displayName,
              });

              leaderboardStore.setLeaderboardActiveValue(
                item.displayName,
                event.detail
              );

              if (browser) {
                const filters = prune($leaderboardStore.activeValues);
                leaderboardStore.socket.emit("getBigNumber", {
                  entityType: EntityType.Table,
                  entityID: $leaderboardStore.activeEntityID,
                  expression: "count(*)",
                  filters,
                });
                $leaderboardStore.availableDimensions.forEach(
                  (dimensionName) => {
                    // invalidate the exiting leaderboard?
                    leaderboardStore.socket.emit("getDimensionLeaderboard", {
                      dimensionName,
                      entityType: EntityType.Table,
                      entityID: $leaderboardStore.activeEntityID,
                      filters,
                    });
                  }
                );
              }
            }}
            activeValues={$leaderboardStore?.activeValues[item.displayName] ||
              []}
            displayName={item.displayName}
            values={item.values}
            referenceValue={referenceValue || 0}
          />
        </div>
      </VirtualizedGrid>
    {/key}
  {/if}
</div>
