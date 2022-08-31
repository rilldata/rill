package com.rilldata.calcite;

import com.fasterxml.jackson.core.JsonProcessingException;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;
import org.apache.calcite.sql.parser.SqlParseException;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import java.util.List;

public class CalciteToolboxTest
{
  @Test
  public void testMigrations() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create view a as select 1", """
          {
            "tables": [
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
    Assertions.assertEquals(1, migrationSteps.size());
  }

  @Test
  public void testMigrationsWithFullName() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create view a as select 1", """
          {
            "tables": [
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
  }

  @Test
  public void testMultipleMigrations() throws SqlParseException, JsonProcessingException
  {
    List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations("create table a (id int) ; create view b as select * from a", """
          {
            "tables": [
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
  }
}
