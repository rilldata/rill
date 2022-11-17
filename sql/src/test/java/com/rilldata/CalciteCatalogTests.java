package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.dialects.Dialects;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.ValidationException;
import org.apache.calcite.util.Litmus;
import org.junit.jupiter.api.Assertions;
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

  @Test
  public void testEmptyCatalog() throws IOException, ValidationException, SqlParseException
  {
    CalciteToolbox calciteToolbox = CalciteToolbox.buildToolbox("{}");
    for (Dialects dialect : Dialects.values()) {
      Assertions.assertEquals("SELECT 1", calciteToolbox.getRunnableQuery("SELECT 1", dialect.getSqlDialect()));
    }
  }

  @ParameterizedTest
  @MethodSource("testQueriesParams")
  public void testQueries(String query, String resultant) throws ValidationException, SqlParseException
  {
    for (Dialects dialect : Dialects.values()) {
      String sql = calciteToolbox.getRunnableQuery(query, dialect.getSqlDialect());
      SqlNode actual = calciteToolbox.parseValidatedSql(sql);
      SqlNode expected = calciteToolbox.parseValidatedSql(resultant);
      boolean equal = SqlNode.equalDeep(actual, expected, Litmus.IGNORE);
      Assertions.assertTrue(equal, String.format("Actual: [%s] \n Expected: [%s]", actual, expected));
    }
  }

  private static Stream<Arguments> testQueriesParams()
  {
    return Stream.of(
        Arguments.of("select * from MV",
            """
                SELECT "heroes"."power", COUNT(DISTINCT "heroes"."name") AS "NAMES"
                FROM "main"."heroes" AS "HEROES"
                GROUP BY "heroes"."power\""""
        ),
        Arguments.of("select \"power\" from MV",
            "SELECT \"heroes\".\"power\" FROM \"main\".\"heroes\" AS \"HEROES\""
        ),
        Arguments.of("select \"power\", names from MV where \"power\" > 100",
            """
                SELECT "heroes"."power", COUNT(DISTINCT "heroes"."name") AS "NAMES"
                FROM "main"."heroes" AS "HEROES"
                WHERE "heroes"."power" > 100
                GROUP BY "heroes"."power\""""
        ),
        Arguments.of("select 1 as foo, 'hello' as bar, h1.id, h1.\"power\", h2.name "
                + "from main.heroes h1 join main.heroes h2 on h1.id = h2.id",
            """
                SELECT 1 AS "FOO", 'hello' AS "BAR", "H1"."id" AS "ID", "H1"."power", "H2"."name" AS "NAME"
                FROM "main"."heroes" AS "H1"
                INNER JOIN "main"."heroes" AS "H2" ON "H1"."id" = "H2"."id\""""
        ),
        // if identifier are not quoted, they are converted to upper case in expected result and deep equal fails
        Arguments.of("select lower(name) from main.heroes",
            "SELECT LOWER(\"heroes\".\"name\") FROM \"main\".\"heroes\" AS HEROES"
        )
    );
  }
}
