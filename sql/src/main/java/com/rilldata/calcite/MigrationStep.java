package com.rilldata.calcite;

public class MigrationStep
{
  public String type;
  public String ddl;

  public static MigrationStep dropEntity(String name, String type) {
    return new MigrationStep("DROP " + type + " " + name);
  }

  public static MigrationStep fromDdl(String ddl) {
    return new MigrationStep(ddl);
  }

  public static MigrationStep insertCatalog(String ddl) {
    MigrationStep migrationStep = new MigrationStep(ddl);
    migrationStep.type = "InsertCatalog";
    return migrationStep;
  }

  public MigrationStep(String ddl)
  {
    type = "ExecuteInfra";
    this.ddl = ddl;
  }

  public String toString() {
    return type + ": " + ddl;
  }
}
