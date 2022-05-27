import { writable, get } from "svelte/store";
import type { Writable } from "svelte/store";
import { produce } from "immer";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { isAnythingSelected } from "./_utils";
import type { Socket } from "socket.io-client";

interface Leaderboard {
  value: number;
  label: string;
}

interface ActiveValues {
  [key: string]: string[]
}

export interface ExploreState {
  activeEntityID: string;
  bigNumber: number;
  referenceValue: number;
  leaderboards: Leaderboard[];
  availableDimensions: string[];
  activeValues: ActiveValues;
}

const initialState: ExploreState = {
  activeEntityID: undefined,
  bigNumber: undefined,
  referenceValue: undefined,
  leaderboards: [],
  availableDimensions: [],
  activeValues: {},
};

function setAvailableDimensions(dimensions = []) {
  return (draft) => {
    draft.availableDimensions = dimensions;
  };
}

function setActiveEntityID(entityID) {
  return (draft) => {
    draft.activeEntityID = entityID;
  };
}

function setBigNumber(bigNumber: number) {
  return (draft) => {
    draft.bigNumber = bigNumber;
  };
}

function setReferenceValue(referenceValue: number) {
  return (draft) => {
    draft.referenceValue = referenceValue;
  };
}

function setLeaderboardActiveValue(dimensionName, dimensionValue) {
  return (draft) => {
    if (!draft.activeValues[dimensionName].includes(dimensionValue)) {
      // add to activeValues
      draft.activeValues[dimensionName] = [
        ...draft.activeValues[dimensionName],
        dimensionValue,
      ];
    } else {
      // remove from activeValues
      draft.activeValues[dimensionName] = draft.activeValues[
        dimensionName
      ]?.filter((b) => b !== dimensionValue);
    }
  };
}

function initializeActiveValues(boards = []) {
  return (draft) => {
    draft.activeValues = boards.reduce((acc, leaderboard) => {
      acc[leaderboard.displayName] = [];
      return acc;
    }, {});
  };
}

/** get RID OF THIS!!! */
function initializeLeaderboardActiveValues(dimensionName) {
  return (draft) => {
    if (!(dimensionName in draft.activeValues)) {
      draft.activeValues[dimensionName] = [];
    }
  };
}

function clearLeaderboards() {
  return (draft) => {
    draft.leaderboards = [];
  };
}

function setDimensionLeaderboard(dimensionName, values) {
  return (draft) => {
    const exists = draft.leaderboards.find(
      (leaderboard) => leaderboard?.displayName === dimensionName
    );
    if (exists) {
      exists.values = values;
      exists.displayName = dimensionName;
    } else
      draft.leaderboards = [
        ...draft.leaderboards,
        { displayName: dimensionName, values },
      ];
  };
}

// handle socket updates.
function initializeSockets(store, socket) {
  socket.on("getAvailableDimensions", ({ dimensions }) => {
    // set availableDimensions

    store.setAvailableDimensions(dimensions);

    const storeValue = get(store) as ExploreState;

    // now, uh, calculate all the dimension leaderboards.
    storeValue.availableDimensions.forEach((dimensionName) => {
      socket.emit("getDimensionLeaderboard", {
        dimensionName,
        entityType: EntityType.Table,
        entityID: storeValue.activeEntityID,
      });
    });
  });
  // receive getDimensionLeaderboard responses.
  socket.on("getDimensionLeaderboard", ({ dimensionName, values }) => {
    store.setDimensionLeaderboard(dimensionName, values);
    // add to the activeValues.
    store.initializeLeaderboardActiveValues(dimensionName);
  });
  // receive bigNumber
  socket.on("getBigNumber", ({ value, filters }) => {
    store.setBigNumber(value);

    if (!isAnythingSelected(filters)) {
      //referenceValue = value;
      store.setReferenceValue(value);
    }
  });
}

const actions = {
  setAvailableDimensions,
  setActiveEntityID,
  setBigNumber,
  setReferenceValue,
  setDimensionLeaderboard,
  initializeActiveValues,
  initializeLeaderboardActiveValues,
  setLeaderboardActiveValue,
  clearLeaderboards,
};

export interface ExploreStore extends Writable<ExploreState> {
  socket: Socket;
  setAvailableDimensions,
  setActiveEntityID,
  setBigNumber,
  setReferenceValue,
  setDimensionLeaderboard,
  initializeActiveValues,
  initializeLeaderboardActiveValues,
  setLeaderboardActiveValue,
  clearLeaderboards,
}

export function createLeaderboardStore(socket): ExploreStore {
  const { subscribe, update } = writable(initialState);

  function dispatch(fcn) {
    if (fcn.constructor.name === "AsyncFunction") {
      // treat as thunk.
      fcn(this, () => get(store));
    } else {
      // treat as plain action.
      update((draft) => produce(draft, fcn));
    }
  }
  // add actionSet.
  const actionSet = Object.entries(actions).reduce((actions, [name, fcn]) => {
    // FIXME: find a better solution than this for typescript.
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    actions[name] = (...args) => dispatch(fcn(...args));
    return actions;
  }, {});

  const store = {
    subscribe,
    ...actionSet,
    socket,
  };
  initializeSockets(store, socket);
  return store as ExploreStore;
}
