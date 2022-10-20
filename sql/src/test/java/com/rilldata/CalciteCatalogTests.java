package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.dialects.Dialects;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.ValidationException;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.io.IOException;
import java.util.stream.Stream;

public class CalciteCatalogTests
{
  static CalciteToolbox calciteToolbox;

  @BeforeAll
  static void setUp() throws IOException
  {
    String catalog = new String(StaticSchemaTest.class.getResourceAsStream("/catalog.json").readAllBytes());
    calciteToolbox = CalciteToolbox.buildToolbox(catalog);
  }

  @ParameterizedTest
  @MethodSource("testQueriesParams")
  public void testQueries(String query) throws ValidationException, SqlParseException
  {
    for(Dialects dialect: Dialects.values()) {
      String sql = calciteToolbox.getRunnableQuery(query, dialect.getSqlDialect());
      System.out.println(sql);
    }
  }

  private static Stream<Arguments> testQueriesParams()
  {
    return Stream.of(
        Arguments.of("select * from MV")
    );
  }

  @Test
  public void testQuery() throws SqlParseException, ValidationException
  {
    String query1 = "select 1 as foo, "
        + "'hello' as bar, h1.\"id\", h1.\"power\", h2.\"name\" "
        + "from heroes h1 join main.heroes h2 on h1.\"id\" = h2.\"id\"";
    String resultantQuery = calciteToolbox.getRunnableQuery(query1, Dialects.DUCKDB.getSqlDialect());
    System.out.println(resultantQuery);
  }
}
