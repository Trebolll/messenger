{{- define "app.mergedValues" -}}
{{- $envName := .Values.ENV_NAME | default "prod" -}}
{{- $common := .Files.Get "values-common.yaml" | fromYaml -}}
{{- if not $common -}}
{{- fail "values-common.yaml is missing or empty" -}}
{{- end -}}
{{- $env := .Files.Get (printf "values-%s.yaml" $envName) | fromYaml -}}
{{- if not $env -}}
{{- fail (printf "values-%s.yaml is missing or empty" $envName) -}}
{{- end -}}
{{- mustMergeOverwrite dict $common $env .Values | toYaml -}}
{{- end -}}
