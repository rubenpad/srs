{{/* Expand the name of the chart. */}}
{{- define "srs.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{/* Create chart name and version as used by the chart label. */}}
{{- define "srs.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Generate the standard labels for kubernetes manifests for this chart */}}
{{- define "srs.labels" }}
    app.kubernetes.io/name: {{ include "srs.name" . }}
    helm.sh/chart: {{ include "srs.chart" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}  
{{- end }}