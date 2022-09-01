package com.rilldata.calcite.models;

import java.io.Serializable;
import java.util.Objects;

public class Artifact implements Serializable
{
  private final ArtifactType type;
  private final String name;
  private final String payload;

  public Artifact(ArtifactType type, String name, String payload)
  {
    this.type = type;
    this.name = name;
    this.payload = payload;
  }

  public ArtifactType getType()
  {
    return type;
  }

  public String getName()
  {
    return name;
  }

  public String getPayload()
  {
    return payload;
  }

  @Override public boolean equals(Object o)
  {
    if (this == o) {
      return true;
    }
    if (o == null || getClass() != o.getClass()) {
      return false;
    }
    Artifact artifact = (Artifact) o;
    return type == artifact.type && name.equals(artifact.name) && payload.equals(artifact.payload);
  }

  @Override public int hashCode()
  {
    return Objects.hash(type, name, payload);
  }
}
