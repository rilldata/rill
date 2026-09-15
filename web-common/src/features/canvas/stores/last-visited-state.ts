// The url search params a canvas was last viewed with, keyed by canvas name.
// Used to restore the last visited state when a canvas is opened without any url params.
// This lives in its own module so that consumers which only need to read or clear it
// (e.g. the embed public API) don't have to import all of canvas-entity.
export const lastVisitedState = new Map<string, string>();
