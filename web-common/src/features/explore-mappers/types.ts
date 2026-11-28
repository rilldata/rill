export interface ExploreAvailabilityResult {
  isAvailable: boolean;
  exploreName?: string;
  displayName?: string;
  error?: string;
}

export interface DashboardSelectionCriteria {
  preferredType?: "recent" | "most_used" | "first_available";
}

export enum ExploreLinkErrorType {
  VALIDATION_ERROR = "VALIDATION_ERROR",
  TRANSFORMATION_ERROR = "TRANSFORMATION_ERROR",
  NETWORK_ERROR = "NETWORK_ERROR",
  PERMISSION_ERROR = "PERMISSION_ERROR",
}

export interface ExploreLinkError {
  type: ExploreLinkErrorType;
  message: string;
  details?: any;
}
