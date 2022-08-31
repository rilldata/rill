package com.rilldata;

import com.rilldata.calcite.DependencyFinder;
import com.rilldata.calcite.generated.RillSqlParserImpl;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParser;
import org.apache.calcite.sql.validate.SqlConformanceEnum;
import org.apache.calcite.util.SourceStringReader;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.List;

public class DependencyFinderTest
{
  @Test
  public void testSanity() throws SqlParseException
  {
    SqlParser.Config parserConfig = SqlParser
        .config()
        .withCaseSensitive(false)
        .withConformance(SqlConformanceEnum.BABEL)
        .withParserFactory(RillSqlParserImpl::new);

    String s = """
        create view a as select 1 union all select a as v from (select * from (select n from d) as k join b on k.a = b.a)
        """;
    SqlNodeList sqlNodes = SqlParser.create(new SourceStringReader(s), parserConfig).parseStmtList();
    List<String> dependencies = sqlNodes.get(0).accept(new DependencyFinder());
    Assertions.assertIterableEquals(Arrays.asList("D", "B"), dependencies);
  }

  @Test
  public void testSanity2() throws SqlParseException
  {
    SqlParser.Config parserConfig = SqlParser
        .config()
        .withCaseSensitive(false)
        .withConformance(SqlConformanceEnum.BABEL)
        .withParserFactory(RillSqlParserImpl::new);

    String s = """
        create table a (a int)
        """;
    SqlNodeList sqlNodes = SqlParser.create(new SourceStringReader(s), parserConfig).parseStmtList();
    List<String> dependencies = sqlNodes.get(0).accept(new DependencyFinder());
    Assertions.assertIterableEquals(Arrays.asList(), dependencies);
  }
}