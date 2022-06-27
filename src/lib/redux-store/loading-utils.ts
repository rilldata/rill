export enum LoadingStatus {
  Idle,
  Loading,
  Error,
}

export function getLoadingState() {
  return { entityStatus: new Map<string, LoadingStatus>() };
}
