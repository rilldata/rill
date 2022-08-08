export enum DatabaseActionQueuePriority {
  TableImport,
  ActiveModel,
  ActiveModelProfile,
  TableProfile,
  ModelExport,
  InactiveModelProfile,
}

export enum ProfilesPriority {
  ActiveModelProfile = 2,
  TableProfile = 3,
  InactiveModelProfile = 5,
}

export enum ProfilesFieldPriority {
  Focused = 0,
  NonFocused = 0.1,
}

export enum ProfilesMetadataPriority {
  SummaryProfile = 0,
  EssentialProfileDetails = 0.01,
  DeeperProfileDetails = 0.02,
}

export function getProfilePriority(
  entity: ProfilesPriority,
  field: ProfilesFieldPriority,
  metadata: ProfilesMetadataPriority
) {
  return entity + field + metadata;
}
