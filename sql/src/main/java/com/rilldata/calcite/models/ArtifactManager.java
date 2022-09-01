package com.rilldata.calcite.models;

import java.util.Set;

public interface ArtifactManager
{
  /**
   * Throws a RuntimeException if artifact of this already exists
   * */
  void saveArtifact(Artifact artifact);

  Artifact getArtifact(ArtifactType artifactType, String name);

  Set<String> getArtifactsOfType(ArtifactType artifactType);
}
