*{{ .Subject }}*

*{{ .Title }}*

{{- if .IsError }}
The alert failed to evaluate on *{{ .ExecutionTimeString }}*. It failed with the following error message: _{{ .ErrorMessage }}_.
{{- else if .IsRecover }}
The alert has recovered on *{{ .ExecutionTimeString }}* from a previous failure.
{{- else if .IsPass }}
The alert has passed on *{{ .ExecutionTimeString }}*.
{{- end }}

<{{ .OpenLink }}|Open in browser>

To edit or unsubscribe from this alert, <{{ .EditLink }}|click here>.

© 2023 Rill Data, Inc
18 Bartol St., San Francisco, CA
<https://www.rilldata.com/contact|Contact us> • <https://bit.ly/3unvA05|Community> • <https://www.rilldata.com/legal/privacy|Privacy Policy>
