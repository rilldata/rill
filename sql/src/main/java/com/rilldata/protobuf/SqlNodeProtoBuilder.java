package com.rilldata.protobuf;

import com.rilldata.calcite.models.SqlCreateMetricsView;
import com.rilldata.calcite.models.SqlCreateSource;
import com.rilldata.protobuf.generated.BasicSqlTypeProto;
import com.rilldata.protobuf.generated.CoercibilityProto;
import com.rilldata.protobuf.generated.IntervalSqlTypeProto;
import com.rilldata.protobuf.generated.RelCrossTypeProto;
import com.rilldata.protobuf.generated.RelDataTypeFieldImplProto;
import com.rilldata.protobuf.generated.RelDataTypeFieldProto;
import com.rilldata.protobuf.generated.RelDataTypeProto;
import com.rilldata.protobuf.generated.RelRecordTypeProto;
import com.rilldata.protobuf.generated.SerializableCharsetProto;
import com.rilldata.protobuf.generated.SqlAlienSystemTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlBasicCallProto;
import com.rilldata.protobuf.generated.SqlBasicTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlCharStringLiteralProto;
import com.rilldata.protobuf.generated.SqlCollationProto;
import com.rilldata.protobuf.generated.SqlCollectionTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlCreateMetricsViewProto;
import com.rilldata.protobuf.generated.SqlCreateSourceProto;
import com.rilldata.protobuf.generated.SqlDataTypeSpecProto;
import com.rilldata.protobuf.generated.SqlDateLiteralProto;
import com.rilldata.protobuf.generated.SqlIdentifierProto;
import com.rilldata.protobuf.generated.SqlIntervalQualifierProto;
import com.rilldata.protobuf.generated.SqlJoinProto;
import com.rilldata.protobuf.generated.SqlKindProto;
import com.rilldata.protobuf.generated.SqlLiteralProto;
import com.rilldata.protobuf.generated.SqlNodeListProto;
import com.rilldata.protobuf.generated.SqlNodeProto;
import com.rilldata.protobuf.generated.SqlNumericLiteralProto;
import com.rilldata.protobuf.generated.SqlOperatorProto;
import com.rilldata.protobuf.generated.SqlOrderByProto;
import com.rilldata.protobuf.generated.SqlParserPosProto;
import com.rilldata.protobuf.generated.SqlRowTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlSelectProto;
import com.rilldata.protobuf.generated.SqlTimeLiteralProto;
import com.rilldata.protobuf.generated.SqlTimestampLiteralProto;
import com.rilldata.protobuf.generated.SqlTypeNameProto;
import com.rilldata.protobuf.generated.SqlTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlUserDefinedTypeNameSpecProto;
import com.rilldata.protobuf.generated.SqlWithItemProto;
import com.rilldata.protobuf.generated.SqlWithProto;
import com.rilldata.protobuf.generated.StructKindProto;
import com.rilldata.protobuf.generated.TimeUnitRangeProto;
import org.apache.calcite.avatica.util.TimeUnitRange;
import org.apache.calcite.rel.type.RelCrossType;
import org.apache.calcite.rel.type.RelDataType;
import org.apache.calcite.rel.type.RelDataTypeField;
import org.apache.calcite.rel.type.RelDataTypeFieldImpl;
import org.apache.calcite.rel.type.RelRecordType;
import org.apache.calcite.rel.type.StructKind;
import org.apache.calcite.sql.SqlAlienSystemTypeNameSpec;
import org.apache.calcite.sql.SqlBasicCall;
import org.apache.calcite.sql.SqlBasicTypeNameSpec;
import org.apache.calcite.sql.SqlCharStringLiteral;
import org.apache.calcite.sql.SqlCollation;
import org.apache.calcite.sql.SqlCollectionTypeNameSpec;
import org.apache.calcite.sql.SqlDataTypeSpec;
import org.apache.calcite.sql.SqlDateLiteral;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlIntervalQualifier;
import org.apache.calcite.sql.SqlJoin;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.SqlNumericLiteral;
import org.apache.calcite.sql.SqlOperator;
import org.apache.calcite.sql.SqlOrderBy;
import org.apache.calcite.sql.SqlRowTypeNameSpec;
import org.apache.calcite.sql.SqlSelect;
import org.apache.calcite.sql.SqlTimeLiteral;
import org.apache.calcite.sql.SqlTimestampLiteral;
import org.apache.calcite.sql.SqlTypeNameSpec;
import org.apache.calcite.sql.SqlUserDefinedTypeNameSpec;
import org.apache.calcite.sql.SqlWith;
import org.apache.calcite.sql.SqlWithItem;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.sql.type.BasicSqlType;
import org.apache.calcite.sql.type.IntervalSqlType;
import org.apache.calcite.sql.type.SqlTypeName;
import org.apache.calcite.sql.validate.SqlValidator;

import javax.annotation.Nullable;
import java.nio.charset.Charset;
import java.util.List;

/**
 * Run `mvn package` to generate the protobuf builder classes in target/generated-sources/annotations folder.
 * It uses following script - proto-gen.sh
 */
public class SqlNodeProtoBuilder
{
  private final SqlNode sqlNode;
  // If type information is not required then this will be null
  @Nullable private final SqlValidator sqlValidator;

  public SqlNodeProtoBuilder(SqlNode sqlNode, @Nullable SqlValidator sqlValidator)
  {
    this.sqlNode = sqlNode;
    this.sqlValidator = sqlValidator;
  }

  public byte[] getProto()
  {
    SqlNodeProto sqlNodeProto = getSqlNodeProto();
    return sqlNodeProto.toByteArray();
  }

  public SqlNodeProto getSqlNodeProto()
  {
    return handleSqlNode(sqlNode);
  }

  private SqlSelectProto handleSqlSelect(SqlSelect sqlSelect)
  {
    SqlSelectProto.Builder sqlSelectBuilder = SqlSelectProto.newBuilder();
    List<SqlNode> operands = sqlSelect.getOperandList();
    // first operand is the keyword list
    SqlNodeList keywordList = operands.get(0) != null ? (SqlNodeList) operands.get(0) : null;
    if (keywordList != null && keywordList.size() > 0) {
      sqlSelectBuilder.setKeywordList(handleSqlNodeList(keywordList));
    }
    // handle select list
    sqlSelectBuilder.setSelectList(handleSqlNodeList(sqlSelect.getSelectList()));
    // handle from
    if (sqlSelect.getFrom() != null) {
      sqlSelectBuilder.setFrom(handleSqlNode(sqlSelect.getFrom()));
    }
    // handle where clause
    if (sqlSelect.getWhere() != null) {
      sqlSelectBuilder.setWhere(handleSqlNode(sqlSelect.getWhere()));
    }
    // handle group by
    if (sqlSelect.getGroup() != null && sqlSelect.getGroup().size() > 0) {
      sqlSelectBuilder.setGroupBy(handleSqlNodeList(sqlSelect.getGroup()));
    }
    // handle having clause
    if (sqlSelect.getHaving() != null) {
      sqlSelectBuilder.setHaving(handleSqlNode(sqlSelect.getHaving()));
    }
    // handle window list
    if (sqlSelect.getWindowList() != null && sqlSelect.getWindowList().size() > 0) {
      sqlSelectBuilder.setWindowDecls(handleSqlNodeList(sqlSelect.getWindowList()));
    }
    // handle order by list
    if (sqlSelect.getOrderList() != null && sqlSelect.getOrderList().size() > 0) {
      sqlSelectBuilder.setOrderBy(handleSqlNodeList(sqlSelect.getOrderList()));
    }
    // handle offset
    if (sqlSelect.getOffset() != null) {
      sqlSelectBuilder.setOffset(handleSqlNode(sqlSelect.getOffset()));
    }
    // handle fetch/limit
    if (sqlSelect.getFetch() != null) {
      sqlSelectBuilder.setFetch(handleSqlNode(sqlSelect.getFetch()));
    }
    if (sqlSelect.getHints() != null && sqlSelect.getHints().size() > 0) {
      sqlSelectBuilder.setHints(handleSqlNodeList(sqlSelect.getHints()));
    }
    // handle pos
    sqlSelectBuilder.setPos(handleParserPos(sqlSelect.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlSelect);
      if (relDataType != null) {
        sqlSelectBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlSelectBuilder.build();
  }

  private SqlOrderByProto handleSqlOrderBy(SqlOrderBy sqlOrderBy)
  {
    SqlOrderByProto.Builder sqlOrderByProtoBuilder = SqlOrderByProto.newBuilder();
    sqlOrderByProtoBuilder.setQuery(handleSqlNode(sqlOrderBy.query));
    sqlOrderByProtoBuilder.setOrderList(handleSqlNodeList(sqlOrderBy.orderList));
    if (sqlOrderBy.offset != null) {
      sqlOrderByProtoBuilder.setOffset(handleSqlNode(sqlOrderBy.offset));
    }
    sqlOrderByProtoBuilder.setFetch(handleSqlNode(sqlOrderBy.fetch));
    sqlOrderByProtoBuilder.setPos(handleParserPos(sqlOrderBy.getParserPosition()));
    return sqlOrderByProtoBuilder.build();
  }

  private SqlNodeListProto handleSqlNodeList(SqlNodeList sqlNodeList)
  {
    SqlNodeListProto.Builder sqlNodeListProtoBuilder = SqlNodeListProto.newBuilder();
    for (SqlNode sqlNode : sqlNodeList.getList()) {
      if (sqlNode != null) {
        sqlNodeListProtoBuilder.addList(handleSqlNode(sqlNode));
      }
    }
    sqlNodeListProtoBuilder.setPos(handleParserPos(sqlNodeList.getParserPosition()));
    return sqlNodeListProtoBuilder.build();
  }

  private SqlWithProto handleSqlWith(SqlWith sqlWith)
  {
    SqlWithProto.Builder sqlWithProtoBuilder = SqlWithProto.newBuilder();
    sqlWithProtoBuilder.setWithList(handleSqlNodeList(sqlWith.withList));
    sqlWithProtoBuilder.setBody(handleSqlNode(sqlWith.body));
    sqlWithProtoBuilder.setPos(handleParserPos(sqlWith.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlWith);
      if (relDataType != null) {
        sqlWithProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlWithProtoBuilder.build();
  }

  private SqlWithItemProto handleSqlWithItem(SqlWithItem sqlWithItem)
  {
    SqlWithItemProto.Builder sqlWithItemProtoBuilder = SqlWithItemProto.newBuilder();
    sqlWithItemProtoBuilder.setName(handleSqlIdentifier(sqlWithItem.name));
    if (sqlWithItem.columnList != null) {
      sqlWithItemProtoBuilder.setColumnList(handleSqlNodeList(sqlWithItem.columnList));
    }
    sqlWithItemProtoBuilder.setQuery(handleSqlNode(sqlWithItem.query));
    sqlWithItemProtoBuilder.setPos(handleParserPos(sqlWithItem.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlWithItem);
      if (relDataType != null) {
        sqlWithItemProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlWithItemProtoBuilder.build();
  }

  private SqlJoinProto handleSqlJoin(SqlJoin sqlJoin)
  {
    SqlJoinProto.Builder sqlJoinProtoBuilder = SqlJoinProto.newBuilder();
    sqlJoinProtoBuilder.setLeft(handleSqlNode(sqlJoin.getLeft()));
    sqlJoinProtoBuilder.setRight(handleSqlNode(sqlJoin.getRight()));
    sqlJoinProtoBuilder.setNatural(handleSqlLiteral(sqlJoin.isNaturalNode()));
    sqlJoinProtoBuilder.setJoinType(handleSqlLiteral(sqlJoin.getJoinTypeNode()));
    sqlJoinProtoBuilder.setConditionType(handleSqlLiteral(sqlJoin.getConditionTypeNode()));
    if (sqlJoin.getCondition() != null) {
      sqlJoinProtoBuilder.setCondition(handleSqlNode(sqlJoin.getCondition()));
    }
    sqlJoinProtoBuilder.setPos(handleParserPos(sqlJoin.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlJoin);
      if (relDataType != null) {
        sqlJoinProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlJoinProtoBuilder.build();
  }

  private SqlBasicCallProto handleSqlBasicCall(SqlBasicCall sqlBasicCall)
  {
    SqlBasicCallProto.Builder sqlBasicCallProtoBuilder = SqlBasicCallProto.newBuilder();
    sqlBasicCallProtoBuilder.setPos(handleParserPos(sqlBasicCall.getParserPosition()));
    if (sqlBasicCall.getFunctionQuantifier() != null) {
      sqlBasicCallProtoBuilder.setFunctionQuantifier(handleSqlLiteral(sqlBasicCall.getFunctionQuantifier()));
    }
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlBasicCall);
      if (relDataType != null) {
        sqlBasicCallProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    sqlBasicCallProtoBuilder.setOperator(handleSqlOperator(sqlBasicCall.getOperator()));
    for (SqlNode operand : sqlBasicCall.getOperandList()) {
      sqlBasicCallProtoBuilder.addOperandList(handleSqlNode(operand));
    }
    return sqlBasicCallProtoBuilder.build();
  }

  private RelDataTypeProto handleRelDataType(RelDataType relDataType)
  {
    if (relDataType instanceof BasicSqlType) {
      return RelDataTypeProto.newBuilder().setBasicSqlTypeProto(handleBasicSqlType((BasicSqlType) relDataType)).build();
    } else if (relDataType instanceof RelRecordType) {
      return RelDataTypeProto.newBuilder().setRelRecordTypeProto(handleRelRecordType((RelRecordType) relDataType))
          .build();
    } else if (relDataType instanceof RelCrossType) {
      return RelDataTypeProto.newBuilder().setRelCrossTypeProto(handleRelCrossType((RelCrossType) relDataType)).build();
    } else if (relDataType instanceof IntervalSqlType) {
      return RelDataTypeProto.newBuilder().setIntervalSqlTypeProto(handleIntervalSqlType((IntervalSqlType) relDataType))
          .build();
    }
    return null;
  }

  private RelCrossTypeProto handleRelCrossType(RelCrossType relCrossType)
  {
    RelCrossTypeProto.Builder relCrossTypeProtoBuilder = RelCrossTypeProto.newBuilder();
    for (RelDataType relDataType : relCrossType.types) {
      relCrossTypeProtoBuilder.addTypes(handleRelDataType(relDataType));
    }
    for (RelDataTypeField relDataTypeField : relCrossType.getFieldList()) {
      relCrossTypeProtoBuilder.addFieldList(handleRelDataTypeFieldProto(relDataTypeField));
    }
    relCrossTypeProtoBuilder.setDigest(relCrossType.getFullTypeString());
    return relCrossTypeProtoBuilder.build();
  }

  private BasicSqlTypeProto handleBasicSqlType(BasicSqlType basicSqlType)
  {
    BasicSqlTypeProto.Builder basicSqlTypeProtoBuilder = BasicSqlTypeProto.newBuilder();
    basicSqlTypeProtoBuilder.setPrecision(basicSqlType.getPrecision());
    basicSqlTypeProtoBuilder.setScale(basicSqlType.getScale());
    basicSqlTypeProtoBuilder.setDigest(basicSqlType.getFullTypeString());
    basicSqlTypeProtoBuilder.setTypeName(handleSqlTypeName(basicSqlType.getSqlTypeName()));
    if (basicSqlType.getCollation() != null) {
      basicSqlTypeProtoBuilder.setCollation(handleSqlCollation(basicSqlType.getCollation()));
    }
    basicSqlTypeProtoBuilder.setIsNullable(basicSqlType.isNullable());
    if (basicSqlType.getFieldList() != null && basicSqlType.getFieldList().size() > 0) {
      for (RelDataTypeField relDataTypeField : basicSqlType.getFieldList()) {
        basicSqlTypeProtoBuilder.addFieldList(handleRelDataTypeFieldProto(relDataTypeField));
      }
    }
    return basicSqlTypeProtoBuilder.build();
  }

  private RelRecordTypeProto handleRelRecordType(RelRecordType relRecordType)
  {
    RelRecordTypeProto.Builder relRecordTypeProtoBuilder = RelRecordTypeProto.newBuilder();
    relRecordTypeProtoBuilder.setKind(handleStructKind(relRecordType.getStructKind()));
    relRecordTypeProtoBuilder.setNullable(relRecordType.isNullable());
    // TODO getFieldMap is protected method in RelRecordType so ignoring it
    for (RelDataTypeField relDataTypeField : relRecordType.getFieldList()) {
      relRecordTypeProtoBuilder.addFieldList(handleRelDataTypeFieldProto(relDataTypeField));
    }
    relRecordTypeProtoBuilder.setDigest(relRecordType.getFullTypeString());
    return relRecordTypeProtoBuilder.build();
  }

  private RelDataTypeFieldProto handleRelDataTypeFieldProto(RelDataTypeField relDataTypeField)
  {
    if (relDataTypeField instanceof RelDataTypeFieldImpl) {
      return RelDataTypeFieldProto.newBuilder()
          .setRelDataTypeFieldImplProto(handleRelDataTypeFieldImplProto((RelDataTypeFieldImpl) relDataTypeField))
          .build();
    }
    return null;
  }

  private RelDataTypeFieldImplProto handleRelDataTypeFieldImplProto(RelDataTypeFieldImpl relDataTypeFieldImpl)
  {
    RelDataTypeFieldImplProto.Builder relDataTypeFieldImplProtoBuilder = RelDataTypeFieldImplProto.newBuilder();
    relDataTypeFieldImplProtoBuilder.setName(relDataTypeFieldImpl.getName());
    relDataTypeFieldImplProtoBuilder.setIndex(relDataTypeFieldImpl.getIndex());
    relDataTypeFieldImplProtoBuilder.setType(handleRelDataType(relDataTypeFieldImpl.getType()));
    return relDataTypeFieldImplProtoBuilder.build();
  }

  private IntervalSqlTypeProto handleIntervalSqlType(IntervalSqlType intervalSqlType)
  {
    IntervalSqlTypeProto.Builder intervalSqlTypeBuilder = IntervalSqlTypeProto.newBuilder();
    // TODO handle intervalSqlTypeBuilder.setTypeSystem()
    intervalSqlTypeBuilder.setIntervalQualifier(handleSqlIntervalQualifier(intervalSqlType.getIntervalQualifier()));
    intervalSqlTypeBuilder.setTypeName(handleSqlTypeName(intervalSqlType.getSqlTypeName()));
    intervalSqlTypeBuilder.setIsNullable(intervalSqlType.isNullable());
    if (intervalSqlType.getFieldList() != null) {
      for (RelDataTypeField relDataTypeField : intervalSqlType.getFieldList()) {
        intervalSqlTypeBuilder.addFieldList(handleRelDataTypeFieldProto(relDataTypeField));
      }
    }
    intervalSqlTypeBuilder.setDigest(intervalSqlType.getFullTypeString());
    return intervalSqlTypeBuilder.build();
  }

  private SqlCollationProto handleSqlCollation(SqlCollation sqlCollation)
  {
    SqlCollationProto.Builder sqlCollationProtoBuilder = SqlCollationProto.newBuilder();
    sqlCollationProtoBuilder.setCollationName(sqlCollation.getCollationName());
    sqlCollationProtoBuilder.setWrappedCharset(handleSerializableCharset(sqlCollation.getCharset()));
    // TODO handle locale, strength
    sqlCollationProtoBuilder.setCoercibility(handleCoercibility(sqlCollation.getCoercibility()));
    return sqlCollationProtoBuilder.build();
  }

  private SerializableCharsetProto handleSerializableCharset(Charset charset)
  {
    SerializableCharsetProto.Builder serializableCharsetProtoBuilder = SerializableCharsetProto.newBuilder();
    serializableCharsetProtoBuilder.setCharsetName(charset.name());
    return serializableCharsetProtoBuilder.build();
  }

  private SqlOperatorProto handleSqlOperator(SqlOperator sqlOperator)
  {
    // TODO handle operator in very generic way as of now, not adding function specific details
    //  as it would involve handle each operator separately and there are lot many operators
    SqlOperatorProto.Builder sqlOperatorProtoBuilder = SqlOperatorProto.newBuilder();
    sqlOperatorProtoBuilder.setName(sqlOperator.getName());
    sqlOperatorProtoBuilder.setKind(handleSqlKind(sqlOperator.getKind()));
    sqlOperatorProtoBuilder.setLeftPrec(sqlOperator.getLeftPrec());
    sqlOperatorProtoBuilder.setRightPrec(sqlOperator.getRightPrec());
    if (sqlValidator != null) {
      if (sqlOperator.getOperandTypeChecker() != null) {
        // default implementation uses operand type checker to get allowed signatures
        String allowedSignatures = sqlOperator.getAllowedSignatures();
        if (allowedSignatures != null) {
          sqlOperatorProtoBuilder.setAllowedSignatures(allowedSignatures);
        } else {
          System.out.println("Found no allowed signatures for operator " + sqlOperator.getName());
        }
      }
    }
    // TODO handle SqlReturnTypeInference, SqlOperandTypeInference and SqlOperandTypeChecker
    return sqlOperatorProtoBuilder.build();
  }

  private SqlIdentifierProto handleSqlIdentifier(SqlIdentifier sqlIdentifier)
  {
    SqlIdentifierProto.Builder sqlIdentifierProto = SqlIdentifierProto.newBuilder();
    for (int i = 0; i < sqlIdentifier.names.size(); i++) {
      sqlIdentifierProto.addNames(sqlIdentifier.names.get(i));
      sqlIdentifierProto.addComponentPositions(handleParserPos(sqlIdentifier.getComponentParserPosition(i)));
    }
    sqlIdentifierProto.setPos(handleParserPos(sqlIdentifier.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlIdentifier);
      if (relDataType != null) {
        sqlIdentifierProto.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    // TODO there is collation property as well but ignoring it for now
    return sqlIdentifierProto.build();
  }

  private SqlLiteralProto handleSqlLiteral(SqlLiteral sqlLiteral)
  {
    SqlLiteralProto.Builder sqlLiteralProtoBuilder = SqlLiteralProto.newBuilder();
    if (sqlLiteral instanceof SqlNumericLiteral) {
      return sqlLiteralProtoBuilder.setSqlNumericLiteralProto(handleSqlNumericLiteral((SqlNumericLiteral) sqlLiteral))
          .build();
    } else if (sqlLiteral instanceof SqlCharStringLiteral) {
      return sqlLiteralProtoBuilder.setSqlCharStringLiteralProto(
          handleSqlCharStringLiteral((SqlCharStringLiteral) sqlLiteral)).build();
    } else if (sqlLiteral instanceof SqlTimestampLiteral) {
      return sqlLiteralProtoBuilder.setSqlTimestampLiteralProto(
          handleSqlTimestampLiteral((SqlTimestampLiteral) sqlLiteral)).build();
    } else if (sqlLiteral instanceof SqlDateLiteral) {
      return sqlLiteralProtoBuilder.setSqlDateLiteralProto(
          handleSqlDateLiteral((SqlDateLiteral) sqlLiteral)).build();
    } else if (sqlLiteral instanceof SqlTimeLiteral) {
      return sqlLiteralProtoBuilder.setSqlTimeLiteralProto(
          handleSqlTimeLiteral((SqlTimeLiteral) sqlLiteral)).build();
    }
    sqlLiteralProtoBuilder.setValue(sqlLiteral.toValue());
    SqlTypeNameProto typeNameProto = handleSqlTypeName(sqlLiteral.getTypeName());
    sqlLiteralProtoBuilder.setTypeName(typeNameProto);
    sqlLiteralProtoBuilder.setPos(handleParserPos(sqlLiteral.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlLiteral);
      if (relDataType != null) {
        sqlLiteralProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlLiteralProtoBuilder.build();
  }

  private SqlNumericLiteralProto handleSqlNumericLiteral(SqlNumericLiteral sqlNumericLiteral)
  {
    SqlNumericLiteralProto.Builder sqlLiteralProtoBuilder = SqlNumericLiteralProto.newBuilder();
    if (sqlNumericLiteral.getPrec() != null) {
      sqlLiteralProtoBuilder.setPrec(sqlNumericLiteral.getPrec());
    }
    if (sqlNumericLiteral.getScale() != null) {
      sqlLiteralProtoBuilder.setScale(sqlNumericLiteral.getScale());
    }
    sqlLiteralProtoBuilder.setValue(sqlNumericLiteral.toValue());
    sqlLiteralProtoBuilder.setIsExact(sqlNumericLiteral.isExact());
    sqlLiteralProtoBuilder.setTypeName(handleSqlTypeName(sqlNumericLiteral.getTypeName()));
    sqlLiteralProtoBuilder.setPos(handleParserPos(sqlNumericLiteral.getParserPosition()));
    if (sqlValidator != null) {
      RelDataType relDataType = sqlValidator.getValidatedNodeTypeIfKnown(sqlNumericLiteral);
      if (relDataType != null) {
        sqlLiteralProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
      }
    }
    return sqlLiteralProtoBuilder.build();
  }

  private SqlCharStringLiteralProto handleSqlCharStringLiteral(SqlCharStringLiteral sqlCharStringLiteral)
  {
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlCharStringLiteral) : null;
    if (relDataType == null) {
      return SqlCharStringLiteralProto.newBuilder().setTypeName(handleSqlTypeName(sqlCharStringLiteral.getTypeName()))
          .setValue(sqlCharStringLiteral.toValue())
          .setPos(handleParserPos(sqlCharStringLiteral.getParserPosition())).build();
    } else {
      return SqlCharStringLiteralProto.newBuilder().setTypeName(handleSqlTypeName(sqlCharStringLiteral.getTypeName()))
          .setValue(sqlCharStringLiteral.toValue()).setTypeInformation(handleRelDataType(relDataType))
          .setPos(handleParserPos(sqlCharStringLiteral.getParserPosition())).build();
    }
  }

  private SqlTimestampLiteralProto handleSqlTimestampLiteral(SqlTimestampLiteral sqlTimestampLiteral)
  {
    SqlTimestampLiteralProto.Builder sqlTimestampLiteralProtoBuilder = SqlTimestampLiteralProto.newBuilder();
    sqlTimestampLiteralProtoBuilder.setPos(handleParserPos(sqlTimestampLiteral.getParserPosition()));
    // TODO sqlTimestampLiteralProtoBuilder.setHasTimeZone() hasTimeZone is not visible
    sqlTimestampLiteralProtoBuilder.setPrecision(sqlTimestampLiteral.getPrec());
    sqlTimestampLiteralProtoBuilder.setTypeName(handleSqlTypeName(sqlTimestampLiteral.getTypeName()));
    sqlTimestampLiteralProtoBuilder.setValue(sqlTimestampLiteral.toFormattedString());
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlTimestampLiteral) : null;
    if (relDataType != null) {
      sqlTimestampLiteralProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlTimestampLiteralProtoBuilder.build();
  }

  private SqlTimeLiteralProto handleSqlTimeLiteral(SqlTimeLiteral sqlTimeLiteral)
  {
    SqlTimeLiteralProto.Builder sqlTimeLiteralProtoBuilder = SqlTimeLiteralProto.newBuilder();
    sqlTimeLiteralProtoBuilder.setPos(handleParserPos(sqlTimeLiteral.getParserPosition()));
    sqlTimeLiteralProtoBuilder.setPrecision(sqlTimeLiteral.getPrec());
    sqlTimeLiteralProtoBuilder.setTypeName(handleSqlTypeName(sqlTimeLiteral.getTypeName()));
    sqlTimeLiteralProtoBuilder.setValue(sqlTimeLiteral.toFormattedString());
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlTimeLiteral) : null;
    if (relDataType != null) {
      sqlTimeLiteralProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlTimeLiteralProtoBuilder.build();
  }

  private SqlDateLiteralProto handleSqlDateLiteral(SqlDateLiteral sqlDateLiteral)
  {
    SqlDateLiteralProto.Builder sqlDateLiteralProtoBuilder = SqlDateLiteralProto.newBuilder();
    sqlDateLiteralProtoBuilder.setPos(handleParserPos(sqlDateLiteral.getParserPosition()));
    sqlDateLiteralProtoBuilder.setPrecision(sqlDateLiteral.getPrec());
    sqlDateLiteralProtoBuilder.setTypeName(handleSqlTypeName(sqlDateLiteral.getTypeName()));
    sqlDateLiteralProtoBuilder.setValue(sqlDateLiteral.toFormattedString());
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlDateLiteral) : null;
    if (relDataType != null) {
      sqlDateLiteralProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlDateLiteralProtoBuilder.build();
  }

  private SqlIntervalQualifierProto handleSqlIntervalQualifier(SqlIntervalQualifier sqlIntervalQualifier)
  {
    SqlIntervalQualifierProto.Builder sqlIntervalQualifierProtoBuilder = SqlIntervalQualifierProto.newBuilder();
    sqlIntervalQualifierProtoBuilder.setPos(handleParserPos(sqlIntervalQualifier.getParserPosition()));
    sqlIntervalQualifierProtoBuilder.setStartPrecision(sqlIntervalQualifier.getStartPrecisionPreservingDefault());
    sqlIntervalQualifierProtoBuilder.setFractionalSecondPrecision(
        sqlIntervalQualifier.getFractionalSecondPrecisionPreservingDefault());
    sqlIntervalQualifierProtoBuilder.setTimeUnitRange(handleTimeUnitRange(sqlIntervalQualifier.timeUnitRange));
    return sqlIntervalQualifierProtoBuilder.build();
  }

  private SqlDataTypeSpecProto handleSqlDataTypeSpec(SqlDataTypeSpec sqlDataTypeSpec)
  {
    SqlDataTypeSpecProto.Builder sqlDataTypeSpecProtoBuilder = SqlDataTypeSpecProto.newBuilder();
    sqlDataTypeSpecProtoBuilder.setTypeNameSpec(handleSqlTypeNameSpec(sqlDataTypeSpec.getTypeNameSpec()));
    // it has Java TimeZone has field, maybe we can just use the display name and set it
    if (sqlDataTypeSpec.getNullable() != null) {
      sqlDataTypeSpecProtoBuilder.setNullable(sqlDataTypeSpec.getNullable());
    }
    sqlDataTypeSpecProtoBuilder.setPos(handleParserPos(sqlDataTypeSpec.getParserPosition()));
    RelDataType relDataType = sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlDataTypeSpec) : null;
    if (relDataType != null) {
      sqlDataTypeSpecProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlDataTypeSpecProtoBuilder.build();
  }

  private SqlTypeNameSpecProto handleSqlTypeNameSpec(SqlTypeNameSpec sqlTypeNameSpec)
  {
    SqlTypeNameSpecProto.Builder sqlTypeNameSpecProtoBuilder = SqlTypeNameSpecProto.newBuilder();
    if (sqlTypeNameSpec instanceof SqlUserDefinedTypeNameSpec) {
      sqlTypeNameSpecProtoBuilder.setSqlUserDefinedTypeNameSpecProto(
          handleSqlUserDefinedTypeNameSpec((SqlUserDefinedTypeNameSpec) sqlTypeNameSpec));
    } else if (sqlTypeNameSpec instanceof SqlRowTypeNameSpec) {
      sqlTypeNameSpecProtoBuilder.setSqlRowTypeNameSpecProto(
          handleSqlRowTypeNameSpec((SqlRowTypeNameSpec) sqlTypeNameSpec));
    } else if (sqlTypeNameSpec instanceof SqlBasicTypeNameSpec) {
      sqlTypeNameSpecProtoBuilder.setSqlBasicTypeNameSpecProto(
          handleSqlBasicTypeNameSpec((SqlBasicTypeNameSpec) sqlTypeNameSpec));
    } else if (sqlTypeNameSpec instanceof SqlCollectionTypeNameSpec) {
      sqlTypeNameSpecProtoBuilder.setSqlCollectionTypeNameSpecProto(
          handleSqlCollectionTypeNameSpec((SqlCollectionTypeNameSpec) sqlTypeNameSpec));
    }
    return sqlTypeNameSpecProtoBuilder.build();
  }

  private SqlUserDefinedTypeNameSpecProto handleSqlUserDefinedTypeNameSpec(
      SqlUserDefinedTypeNameSpec sqlUserDefinedTypeNameSpec
  )
  {
    SqlUserDefinedTypeNameSpecProto.Builder sqlUserDefinedTypeNameSpecProtoBuilder = SqlUserDefinedTypeNameSpecProto.newBuilder();
    sqlUserDefinedTypeNameSpecProtoBuilder.setTypeName(handleSqlIdentifier(sqlUserDefinedTypeNameSpec.getTypeName()));
    sqlUserDefinedTypeNameSpecProtoBuilder.setPos(handleParserPos(sqlUserDefinedTypeNameSpec.getParserPos()));
    return sqlUserDefinedTypeNameSpecProtoBuilder.build();
  }

  private SqlRowTypeNameSpecProto handleSqlRowTypeNameSpec(SqlRowTypeNameSpec sqlRowTypeNameSpec)
  {
    SqlRowTypeNameSpecProto.Builder sqlRowTypeNameSpecProtoBuilder = SqlRowTypeNameSpecProto.newBuilder();
    for (SqlIdentifier fieldName : sqlRowTypeNameSpec.getFieldNames()) {
      sqlRowTypeNameSpecProtoBuilder.addFieldNames(handleSqlIdentifier(fieldName));
    }
    for (SqlDataTypeSpec fieldType : sqlRowTypeNameSpec.getFieldTypes()) {
      sqlRowTypeNameSpecProtoBuilder.addFieldTypes(handleSqlDataTypeSpec(fieldType));
    }
    sqlRowTypeNameSpecProtoBuilder.setTypeName(handleSqlIdentifier(sqlRowTypeNameSpec.getTypeName()));
    sqlRowTypeNameSpecProtoBuilder.setPos(handleParserPos(sqlRowTypeNameSpec.getParserPos()));
    return sqlRowTypeNameSpecProtoBuilder.build();
  }

  private SqlBasicTypeNameSpecProto handleSqlBasicTypeNameSpec(SqlBasicTypeNameSpec sqlBasicTypeNameSpec)
  {
    SqlBasicTypeNameSpecProto.Builder sqlBasicTypeNameSpecProtoBuilder = SqlBasicTypeNameSpecProto.newBuilder();
    if (sqlBasicTypeNameSpec instanceof SqlAlienSystemTypeNameSpec) {
      sqlBasicTypeNameSpecProtoBuilder.setSqlAlienSystemTypeNameSpecProto(
          handleSqlAlienSystemTypeNameSpec((SqlAlienSystemTypeNameSpec) sqlBasicTypeNameSpec));
    } else {
      // sqlTypeName does not have any getter in SqlBasicTypeNameSpec so ignoring it, name of sqlTypeName is set in typeName
      sqlBasicTypeNameSpecProtoBuilder.setPrecision(sqlBasicTypeNameSpec.getPrecision());
      sqlBasicTypeNameSpecProtoBuilder.setScale(sqlBasicTypeNameSpec.getScale());
      if (sqlBasicTypeNameSpec.getCharSetName() != null) {
        sqlBasicTypeNameSpecProtoBuilder.setCharSetName(sqlBasicTypeNameSpec.getCharSetName());
      }
      sqlBasicTypeNameSpecProtoBuilder.setTypeName(handleSqlIdentifier(sqlBasicTypeNameSpec.getTypeName()));
      sqlBasicTypeNameSpecProtoBuilder.setPos(handleParserPos(sqlBasicTypeNameSpec.getParserPos()));
    }
    return sqlBasicTypeNameSpecProtoBuilder.build();
  }

  private SqlAlienSystemTypeNameSpecProto handleSqlAlienSystemTypeNameSpec(
      SqlAlienSystemTypeNameSpec sqlAlienSystemTypeNameSpec
  )
  {
    SqlAlienSystemTypeNameSpecProto.Builder sqlAlienSystemTypeNameSpecProtoBuilder = SqlAlienSystemTypeNameSpecProto.newBuilder();
    // typeAlias is private in SqlAlienSystemTypeNameSpec so essentially SqlAlienSystemTypeNameSpec is same as SqlBasicTypeNameSpec
    sqlAlienSystemTypeNameSpecProtoBuilder.setPrecision(sqlAlienSystemTypeNameSpec.getPrecision());
    sqlAlienSystemTypeNameSpecProtoBuilder.setScale(sqlAlienSystemTypeNameSpec.getScale());
    if (sqlAlienSystemTypeNameSpec.getCharSetName() != null) {
      sqlAlienSystemTypeNameSpecProtoBuilder.setCharSetName(sqlAlienSystemTypeNameSpec.getCharSetName());
    }
    sqlAlienSystemTypeNameSpecProtoBuilder.setTypeName(handleSqlIdentifier(sqlAlienSystemTypeNameSpec.getTypeName()));
    sqlAlienSystemTypeNameSpecProtoBuilder.setPos(handleParserPos(sqlAlienSystemTypeNameSpec.getParserPos()));
    return sqlAlienSystemTypeNameSpecProtoBuilder.build();
  }

  private SqlCollectionTypeNameSpecProto handleSqlCollectionTypeNameSpec(
      SqlCollectionTypeNameSpec sqlCollectionTypeNameSpec
  )
  {
    SqlCollectionTypeNameSpecProto.Builder sqlCollectionTypeNameSpecProtoBuilder = SqlCollectionTypeNameSpecProto.newBuilder();
    sqlCollectionTypeNameSpecProtoBuilder.setElementTypeName(
        handleSqlTypeNameSpec(sqlCollectionTypeNameSpec.getElementTypeName()));
    // collectionTypeName does not have any getter in SqlCollectionTypeNameSpec so ignoring it
    sqlCollectionTypeNameSpecProtoBuilder.setTypeName(handleSqlIdentifier(sqlCollectionTypeNameSpec.getTypeName()));
    sqlCollectionTypeNameSpecProtoBuilder.setPos(handleParserPos(sqlCollectionTypeNameSpec.getParserPos()));
    return sqlCollectionTypeNameSpecProtoBuilder.build();
  }

  private SqlCreateSourceProto handleSqlCreateSource(SqlCreateSource sqlCreateSource)
  {
    SqlCreateSourceProto.Builder sqlCreateSourceProtoBuilder = SqlCreateSourceProto.newBuilder();
    sqlCreateSourceProtoBuilder.setName(handleSqlIdentifier(sqlCreateSource.name));
    sqlCreateSourceProtoBuilder.putAllProperties(sqlCreateSource.properties);
    sqlCreateSourceProtoBuilder.setReplace(sqlCreateSource.getReplace());
    sqlCreateSourceProtoBuilder.setIfNotExists(sqlCreateSource.ifNotExists);
    sqlCreateSourceProtoBuilder.setOperator(handleSqlOperator(sqlCreateSource.getOperator()));
    sqlCreateSourceProtoBuilder.setPos(handleParserPos(sqlCreateSource.getParserPosition()));
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlCreateSource) : null;
    if (relDataType != null) {
      sqlCreateSourceProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlCreateSourceProtoBuilder.build();
  }

  private SqlCreateMetricsViewProto handleSqlCreateMetricsView(SqlCreateMetricsView sqlCreateMetricsView)
  {
    SqlCreateMetricsViewProto.Builder sqlCreateMetricsViewProtoBuilder = SqlCreateMetricsViewProto.newBuilder();
    sqlCreateMetricsViewProtoBuilder.setName(handleSqlIdentifier(sqlCreateMetricsView.name));
    sqlCreateMetricsViewProtoBuilder.setDimensions(handleSqlNodeList(sqlCreateMetricsView.dimensions));
    sqlCreateMetricsViewProtoBuilder.setMeasures(handleSqlNodeList(sqlCreateMetricsView.measures));
    sqlCreateMetricsViewProtoBuilder.setFrom(handleSqlNode(sqlCreateMetricsView.from));
    sqlCreateMetricsViewProtoBuilder.setReplace(sqlCreateMetricsView.getReplace());
    sqlCreateMetricsViewProtoBuilder.setIfNotExists(sqlCreateMetricsView.ifNotExists);
    sqlCreateMetricsViewProtoBuilder.setOperator(handleSqlOperator(sqlCreateMetricsView.getOperator()));
    RelDataType relDataType =
        sqlValidator != null ? sqlValidator.getValidatedNodeTypeIfKnown(sqlCreateMetricsView) : null;
    if (relDataType != null) {
      sqlCreateMetricsViewProtoBuilder.setTypeInformation(handleRelDataType(relDataType));
    }
    return sqlCreateMetricsViewProtoBuilder.build();
  }

  private SqlTypeNameProto handleSqlTypeName(SqlTypeName sqlTypeName)
  {
    return SqlTypeNameProto.valueOf(sqlTypeName.getClass().getSimpleName() + "Proto_" + sqlTypeName.name() + "_");
  }

  private SqlKindProto handleSqlKind(SqlKind sqlKind)
  {
    return SqlKindProto.valueOf(sqlKind.getClass().getSimpleName() + "Proto_" + sqlKind.name() + "_");
  }

  private CoercibilityProto handleCoercibility(SqlCollation.Coercibility coercibility)
  {
    return CoercibilityProto.valueOf(coercibility.getClass().getSimpleName() + "Proto_" + coercibility.name() + "_");
  }

  private StructKindProto handleStructKind(StructKind structKind)
  {
    return StructKindProto.valueOf(structKind.getClass().getSimpleName() + "Proto_" + structKind.name() + "_");
  }

  private TimeUnitRangeProto handleTimeUnitRange(TimeUnitRange timeUnitRange)
  {
    return TimeUnitRangeProto.valueOf(timeUnitRange.getClass().getSimpleName() + "Proto_" + timeUnitRange.name() + "_");
  }

  private SqlParserPosProto handleParserPos(SqlParserPos pos)
  {
    return SqlParserPosProto.newBuilder().setLineNumber(pos.getLineNum()).setColumnNumber(pos.getColumnNum())
        .setEndLineNumber(pos.getEndLineNum()).setEndColumnNumber(pos.getEndColumnNum()).build();
  }

  private SqlNodeProto handleSqlNode(SqlNode sqlNode)
  {
    if (sqlNode instanceof SqlIdentifier) {
      return SqlNodeProto.newBuilder().setSqlIdentifierProto(handleSqlIdentifier((SqlIdentifier) sqlNode)).build();
    } else if (sqlNode instanceof SqlBasicCall) {
      return SqlNodeProto.newBuilder().setSqlBasicCallProto(handleSqlBasicCall((SqlBasicCall) sqlNode)).build();
    } else if (sqlNode instanceof SqlLiteral) {
      return SqlNodeProto.newBuilder().setSqlLiteralProto(handleSqlLiteral((SqlLiteral) sqlNode)).build();
    } else if (sqlNode instanceof SqlSelect) {
      return SqlNodeProto.newBuilder().setSqlSelectProto(handleSqlSelect((SqlSelect) sqlNode)).build();
    } else if (sqlNode instanceof SqlOrderBy) {
      return SqlNodeProto.newBuilder().setSqlOrderByProto(handleSqlOrderBy((SqlOrderBy) sqlNode)).build();
    } else if (sqlNode instanceof SqlNodeList) {
      return SqlNodeProto.newBuilder().setSqlNodeListProto(handleSqlNodeList((SqlNodeList) sqlNode)).build();
    } else if (sqlNode instanceof SqlWith) {
      return SqlNodeProto.newBuilder().setSqlWithProto(handleSqlWith((SqlWith) sqlNode)).build();
    } else if (sqlNode instanceof SqlWithItem) {
      return SqlNodeProto.newBuilder().setSqlWithItemProto(handleSqlWithItem((SqlWithItem) sqlNode)).build();
    } else if (sqlNode instanceof SqlJoin) {
      return SqlNodeProto.newBuilder().setSqlJoinProto(handleSqlJoin((SqlJoin) sqlNode)).build();
    } else if (sqlNode instanceof SqlIntervalQualifier) {
      return SqlNodeProto.newBuilder()
          .setSqlIntervalQualifierProto(handleSqlIntervalQualifier((SqlIntervalQualifier) sqlNode)).build();
    } else if (sqlNode instanceof SqlCreateSource) {
      return SqlNodeProto.newBuilder()
          .setSqlCreateSourceProto(handleSqlCreateSource((SqlCreateSource) sqlNode)).build();
    } else if (sqlNode instanceof SqlCreateMetricsView) {
      return SqlNodeProto.newBuilder()
          .setSqlCreateMetricsViewProto(handleSqlCreateMetricsView((SqlCreateMetricsView) sqlNode)).build();
    } else if (sqlNode instanceof SqlDataTypeSpec) {
      return SqlNodeProto.newBuilder().setSqlDataTypeSpecProto(handleSqlDataTypeSpec((SqlDataTypeSpec) sqlNode))
          .build();
    } else {
      return null;
    }
  }
}
