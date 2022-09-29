package com.rilldata.calcite;

import com.fasterxml.jackson.core.JsonProcessingException;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;
import org.apache.calcite.sql.parser.SqlParseException;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;
import static org.assertj.core.api.Assertions.*;

import java.util.List;

public class CalciteToolboxTest
{
  @Test
  public void testMigrations() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create view a as select 1", """
          {
            "entities": [
              {
                "name": "b",
                "columns": [
                  {
                    "name": "a",
                    "type": "int"
                  }
                ],
                "ddl": "create table b (a int)"
              }
            ]
          }
          """, PostgresqlSqlDialect.DEFAULT);
    Assertions.assertEquals("DROP TABLE \"B\"", migrationSteps.get(0).ddl);
    Assertions.assertEquals("CREATE VIEW \"A\" AS\nSELECT 1", migrationSteps.get(1).ddl);
    Assertions.assertEquals(2, migrationSteps.size());
  }

  @Test
  public void testMigrationsWithArtifacts() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations(
        """
            create view a as select 'a' as d, 2 as m ;
            create metrics view b dimensions d measures count(m) from a
        """,
        """
          {
            "entities": [
              {
                "name": "b",
                "columns": [
                  {
                    "name": "a",
                    "type": "int"
                  }
                ],
                "ddl": "create table b (a int)"
              }
            ]
          }
        """, PostgresqlSqlDialect.DEFAULT);
    assertThat(migrationSteps).extracting("type", "ddl").containsOnly(
        tuple("ExecuteInfra", "DROP TABLE \"B\""),
        tuple("ExecuteInfra", "CREATE VIEW \"A\" AS\nSELECT 'a' AS \"D\", 2 AS \"M\""),
        tuple("InsertCatalog", "CREATE METRICS VIEW \"B\" DIMENSIONS \"D\" MEASURES COUNT(\"M\") FROM \"A\""));
  }

  @Test
  public void testMigrationsWithFullName() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create view a as select 1", """
          {
            "entities": [
              {
                "name": "b",
                "columns": [
                  {
                    "name": "a",
                    "type": "int"
                  }
                ],
                "ddl": "create table main.b (a int)"
              }
            ]
          }
          """, PostgresqlSqlDialect.DEFAULT);
    Assertions.assertEquals("DROP TABLE \"MAIN\".\"B\"", migrationSteps.get(0).ddl);
    Assertions.assertEquals("CREATE VIEW \"A\" AS\nSELECT 1", migrationSteps.get(1).ddl);
  }

  @Test
  public void testMultipleMigrations() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create table a (id int) ; create view b as select * from a", """
          {
            "entities": [
              {
                "name": "a",
                "columns": [
                  {
                    "name": "id",
                    "type": "int"
                  }
                ],
                "ddl": "create table a (id int)"
              },
              {
                "name": "b",
                "columns": [
                  {
                    "name": "id",
                    "type": "int"
                  }
                ],
                "ddl": "create view b as select * from a"
              }
            ]
          }
          """, PostgresqlSqlDialect.DEFAULT);
    Assertions.assertEquals("DROP TABLE \"A\"", migrationSteps.get(0).ddl);
    Assertions.assertEquals("DROP VIEW \"B\"", migrationSteps.get(1).ddl);
    Assertions.assertEquals("CREATE TABLE \"A\" (\"ID\" INTEGER)", migrationSteps.get(2).ddl);
    Assertions.assertEquals("CREATE VIEW \"B\" AS\nSELECT *\nFROM \"A\"", migrationSteps.get(3).ddl);
  }
}
