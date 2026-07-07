{{/* Chart name */}}
{{- define "uptimehub-agent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Fully qualified app name */}}
{{- define "uptimehub-agent.fullname" -}}
{{- if contains .Chart.Name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/* Common labels */}}
{{- define "uptimehub-agent.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
app.kubernetes.io/name: {{ include "uptimehub-agent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/* Selector labels */}}
{{- define "uptimehub-agent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "uptimehub-agent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* Service account name */}}
{{- define "uptimehub-agent.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "uptimehub-agent.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/* Secret name holding the enrollment token */}}
{{- define "uptimehub-agent.secretName" -}}
{{- if .Values.master.existingSecret }}
{{- .Values.master.existingSecret }}
{{- else }}
{{- include "uptimehub-agent.fullname" . }}
{{- end }}
{{- end }}
