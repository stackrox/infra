{{- define "require-file" }}
  {{- $context := (last .) -}}
  {{- $filename := (first .) -}}
  {{- $full_filename := (printf "configuration/%s/%s" (required "A valid .Values.environment entry is required!" $context.Values.environment) $filename) -}}
  {{- if not ($context.Files.Get $full_filename) -}}
    {{- fail (printf "Failed to locate the file %q." $full_filename) -}}
  {{- end -}}
  {{ printf "%s" ($context.Files.Get $full_filename) }}
{{- end }}

{{- define "docker-io-pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.pullSecrets.docker.registry (printf "%s:%s" .Values.pullSecrets.docker.username .Values.pullSecrets.docker.password | b64enc) | b64enc }}
{{- end }}

{{- define "stackrox-io-pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.pullSecrets.stackrox.registry (printf "%s:%s" .Values.pullSecrets.stackrox.username .Values.pullSecrets.stackrox.password | b64enc) | b64enc }}
{{- end }}


{{- define "pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .registry (printf "%s:%s" .username .password | b64enc) | b64enc }}
{{- end }}