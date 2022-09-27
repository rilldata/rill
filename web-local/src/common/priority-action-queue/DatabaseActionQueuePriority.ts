export enum DatabaseActionQueuePriority {
  TableImport = 0,
  ActiveModel = 1,
  ActiveModelProfile = 2,
  TableProfile = 3,
  ModelExport = 4,
  InactiveModelProfile = 5,
}

export enum DatabaseProfilesFieldPriority {
  Focused = 0,
  NonFocused = 0.1,
}

export enum DatabaseProfilesMetadataPriority {
  SummaryProfileDetails = 0,
  EssentialProfileDetails = 0.01,
  DeeperProfileDetails = 0.02,
}

export function getProfilePriority(
  entityType: DatabaseActionQueuePriority,
  fieldType: DatabaseProfilesFieldPriority,
  metadataType: DatabaseProfilesMetadataPriority
) {
  return entityType + fieldType + metadataType;
}

export enum MetadataPriority {
  Summary = "_summary",
  Essential = "_essential",
  Deeper = "_deeper",
}

export const ProfileMetadataPriorityMap = {
  [MetadataPriority.Summary]:
    DatabaseProfilesMetadataPriority.SummaryProfileDetails,
  [MetadataPriority.Essential]:
    DatabaseProfilesMetadataPriority.EssentialProfileDetails,
  [MetadataPriority.Deeper]:
    DatabaseProfilesMetadataPriority.DeeperProfileDetails,
};
