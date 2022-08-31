package com.rilldata.calcite;

public class MigrationStep
{
  public String ddl;

  public MigrationStep(String name, String type)
  {
    ddl = "DROP " + type + " " + name;
  }
}
