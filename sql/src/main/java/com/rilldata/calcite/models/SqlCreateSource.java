package com.rilldata.calcite.models;

import com.rilldata.calcite.generated.ParseException;
import org.apache.calcite.sql.SqlCreate;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlOperator;
import org.apache.calcite.sql.SqlSpecialOperator;
import org.apache.calcite.sql.SqlWriter;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.util.ImmutableNullableList;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class SqlCreateSource extends SqlCreate
{
  public final SqlIdentifier name;
  public final Map<String, String> properties;

  private static final SqlOperator OPERATOR =
      new SqlSpecialOperator("CREATE SOURCE", SqlKind.OTHER);

  public SqlCreateSource(SqlParserPos pos, SqlIdentifier name, Map<SqlNode, SqlNode> properties) throws ParseException
  {
    super(OPERATOR, pos, false, false);
    this.name = name;
    this.properties = new HashMap<>();
    for (Map.Entry<SqlNode, SqlNode> entry : properties.entrySet()) {
      String key = ((SqlLiteral) entry.getKey()).toValue();
      String value = ((SqlLiteral) entry.getValue()).toValue();
      if (key == null || value == null) {
        throw new ParseException(String.format("Either key or value is null for property %s", entry));
      }
      this.properties.put(key.toLowerCase(), value.toLowerCase());
    }
  }

  @Override public void unparse(SqlWriter writer, int leftPrec, int rightPrec)
  {
    writer.keyword("CREATE SOURCE");
    name.unparse(writer, leftPrec, rightPrec);
    writer.keyword("WITH");
    writer.keyword("(");
    for (Map.Entry<String, String> entry : properties.entrySet()) {
      writer.newlineAndIndent();
      writer.literal("'" + entry.getKey() + "'");
      writer.keyword("=");
      writer.literal("'" + entry.getValue() + "'");
      writer.keyword(",");
    }
    writer.newlineAndIndent();
    writer.keyword(")");
  }

  @Override
  public List<SqlNode> getOperandList()
  {
    return ImmutableNullableList.of(name);
  }
}
