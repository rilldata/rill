export enum ValidationState {
  OK = "OK",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export enum SourceModelValidationStatus {
  OK = "OK",
  // No source model selected.
  EMPTY = "EMPTY",
  // Source model query is invalid.
  INVALID = "INVALID",
  // Selected source model is no longer present.
  MISSING = "MISSING",
}
