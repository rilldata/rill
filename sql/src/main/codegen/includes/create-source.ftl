SqlCreateSource SqlCreateSource(Span s, boolean replace) :
{
    final SqlIdentifier id;
    final Map<SqlNode, SqlNode> map;
}
{
    <SOURCE> id = SimpleIdentifier()
    <WITH> <LPAREN>
    map = Properties()
    <RPAREN>
    {
      return new SqlCreateSource(s.end(this), id, map);
    }
}

Map<SqlNode, SqlNode> Properties() :
{
    final Map<SqlNode, SqlNode> props = new HashMap<SqlNode, SqlNode>();
    SqlNode key;
    SqlNode value;
}
{
    key = StringLiteral() <EQ> value = StringLiteral()
    {
      props.put(key, value);
    }
    (
      LOOKAHEAD(2)
      <COMMA> key = StringLiteral() <EQ> value = StringLiteral()
      {
          props.put(key, value);
      }
    )*
    [<COMMA>]
    {
        return props;
    }
}
