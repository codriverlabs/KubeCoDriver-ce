{{/*
Expand the name of the chart.
*/}}
{{- define "kubecodriver-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "kubecodriver-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kubecodriver-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubecodriver-operator.labels" -}}
helm.sh/chart: {{ include "kubecodriver-operator.chart" . }}
{{ include "kubecodriver-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubecodriver-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubecodriver-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller image
*/}}
{{- define "kubecodriver-operator.controller.image" -}}
{{- printf "%s/kubecodriver-controller:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Collector image  
*/}}
{{- define "kubecodriver-operator.collector.image" -}}
{{- printf "%s/kubecodriver-collector:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Aperf image
*/}}
{{- define "kubecodriver-operator.aperf.image" -}}
{{- printf "%s/kubecodriver-aperf:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Tcpdump image
*/}}
{{- define "kubecodriver-operator.tcpdump.image" -}}
{{- printf "%s/kubecodriver-tcpdump:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Chaos image
*/}}
{{- define "kubecodriver-operator.chaos.image" -}}
{{- printf "%s/kubecodriver-chaos:%s" .Values.global.registry.repository .Values.global.version }}
{{- end }}

{{/*
Namespace
*/}}
{{- define "kubecodriver-operator.namespace" -}}
{{- default "kubecodriver-system" .Values.global.namespace }}
{{- end }}

{{/*
Image pull secrets - only include if not using IRSA for ECR
*/}}
{{- define "kubecodriver-operator.imagePullSecrets" -}}
{{- if and (eq .Values.global.registry.type "ecr") (not .Values.ecr.useIRSA) }}
{{- if .Values.ecr.secretName }}
- name: {{ .Values.ecr.secretName }}
{{- end }}
{{- else if .Values.global.imagePullSecrets }}
{{- range .Values.global.imagePullSecrets }}
- name: {{ . }}
{{- end }}
{{- end }}
{{- end }}
