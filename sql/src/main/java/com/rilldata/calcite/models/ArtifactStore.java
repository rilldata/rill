package com.rilldata.calcite.models;

import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

public class ArtifactStore
{
  public static final Set<String> EMPTY_SET = new HashSet<>(0);
  Map<ArtifactType, Map<String, Artifact>> artifacts = new HashMap<>();

  public ArtifactStore() {}

  public ArtifactStore(List<Artifact> artifacts)
  {
    if (artifacts != null) {
      for (Artifact artifact : artifacts) {
        saveArtifact(artifact);
      }
    }
  }

  public void saveArtifact(Artifact artifact)
  {
    artifacts.computeIfAbsent(artifact.getType(), type -> new HashMap<>());
    artifacts.get(artifact.getType()).put(artifact.getName().toLowerCase(), artifact);
  }

  public Artifact getArtifact(ArtifactType artifactType, String name)
  {
    return artifacts.get(artifactType) == null ? null : artifacts.get(artifactType).get(name.toLowerCase());
  }

  public Set<String> getArtifactsOfType(ArtifactType artifactType)
  {
    return artifacts.get(artifactType) == null ? EMPTY_SET : artifacts.get(artifactType).keySet();
  }
}
