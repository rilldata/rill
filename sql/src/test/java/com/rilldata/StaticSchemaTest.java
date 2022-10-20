package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.dialects.Dialects;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.ValidationException;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import java.io.IOException;

public class StaticSchemaTest
{
  @Test
  public void testSanity() throws ValidationException, SqlParseException, IOException
  {
    CalciteToolbox calciteToolbox = new CalciteToolbox(
        new StaticSchemaProvider(new String(StaticSchemaTest.class.getResourceAsStream("/schema.json").readAllBytes())),
        null
    );
    String runnableQuery = calciteToolbox.getRunnableQuery("select \"name\" from \"main\".\"heroes\"",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");
  }

  @Test
  public void testTypes() throws ValidationException, SqlParseException, IOException
  {
    CalciteToolbox calciteToolbox = new CalciteToolbox(
        new StaticSchemaProvider(new String(StaticSchemaTest.class.getResourceAsStream("/schema.json").readAllBytes())),
        null
    );

    String runnableQuery = calciteToolbox.getRunnableQuery("select lower(\"name\") from \"main\".\"heroes\"",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");

    runnableQuery = calciteToolbox.getRunnableQuery(
        "select \"power\" from \"main\".\"heroes\" where \"heroes\".\"power\" > 10000.001",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");

    runnableQuery = calciteToolbox.getRunnableQuery(
        "insert into \"main\".\"heroes\" (\"id\", \"name\", \"power\") values (1, 'Superman', 100.001)",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("insert"), "no insert");

    runnableQuery = calciteToolbox.getRunnableQuery(
        "insert into \"main\".\"heroes\" (\"id\", \"name\", \"power\") values (1, 'Superman', 100000)",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("insert"), "no insert");
  }

  @Test
  public void testTypes2() throws ValidationException, SqlParseException, IOException
  {
    CalciteToolbox calciteToolbox = new CalciteToolbox(
        new StaticSchemaProvider(new String(StaticSchemaTest.class.getResourceAsStream("/schema.json").readAllBytes())),
        null
    );
    String runnableQuery = calciteToolbox.getRunnableQuery("select lower(\"name\") from \"main\".\"heroes\"",
        Dialects.DUCKDB.getSqlDialect()
    );
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");
  }
}
