SqlCreateSource SqlCreateSource(Span s, boolean replace) :
{
    final SqlIdentifier id;
    final Map<SqlNode, SqlNode> map;
}
{
    <SOURCE> id = SimpleIdentifier()
    <WITH>
    (
      <LPAREN> map = Properties() <RPAREN>
    |
      map = Properties()
    )
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
    (LOOKAHEAD(StringLiteral()) key = StringLiteral() | key = SimpleIdentifier())  <EQ> value = StringLiteral()
    {
      props.put(key, value);
    }
    (
      LOOKAHEAD(2)
      <COMMA> (LOOKAHEAD(StringLiteral()) key = StringLiteral() | key = SimpleIdentifier()) <EQ> value = StringLiteral()
      {
          props.put(key, value);
      }
    )*
    [<COMMA>]
    {
        return props;
    }
}
