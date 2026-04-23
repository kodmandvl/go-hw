{{/*
  Общие метки для согласованного выбора подов сервисами и для отладки в кластере.
*/}}
{{- define "calendar-chart.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/name: {{ .Chart.Name | quote }}
{{- end }}
