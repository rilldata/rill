package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;
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
        PostgresqlSqlDialect.DEFAULT,
        null
    );
    String runnableQuery = calciteToolbox.getRunnableQuery("select \"name\" from \"main\".\"heroes\"");
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");
  }

  @Test
  public void testTypes() throws ValidationException, SqlParseException, IOException
  {
    CalciteToolbox calciteToolbox = new CalciteToolbox(
        new StaticSchemaProvider(new String(StaticSchemaTest.class.getResourceAsStream("/schema.json").readAllBytes())),
        PostgresqlSqlDialect.DEFAULT,
        null
    );

    String runnableQuery = calciteToolbox.getRunnableQuery("select lower(\"id\") from \"main\".\"heroes\"");
    System.out.println(runnableQuery);
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");

    runnableQuery = calciteToolbox.getRunnableQuery("select \"power\" from \"main\".\"heroes\" where \"heroes\".\"power\" > 10000.001");
    System.out.println(runnableQuery);
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");

    runnableQuery = calciteToolbox.getRunnableQuery("insert into \"main\".\"heroes\" (\"id\", \"name\", \"power\") values (1, 'Superman', 100.001)");
    System.out.println(runnableQuery);
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("insert"), "no insert");

    runnableQuery = calciteToolbox.getRunnableQuery("insert into \"main\".\"heroes\" (\"id\", \"name\", \"power\") values (1, 'Superman', 100000)");
    System.out.println(runnableQuery);
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("insert"), "no insert");
  }

  @Test
  public void testTypes2() throws ValidationException, SqlParseException, IOException
  {
    CalciteToolbox calciteToolbox = new CalciteToolbox(
        new StaticSchemaProvider(new String(StaticSchemaTest.class.getResourceAsStream("/schema.json").readAllBytes())),
        PostgresqlSqlDialect.DEFAULT,
        null
    );

    String runnableQuery = calciteToolbox.getRunnableQuery("select lower(\"id\") from \"main\".\"heroes\"");
    System.out.println(runnableQuery);
    Assertions.assertTrue(runnableQuery.toLowerCase().contains("select"), "no select");
  }
}
