{{- define "docker-io-pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.pullSecrets.docker.registry (printf "%s:%s" .Values.pullSecrets.docker.username .Values.pullSecrets.docker.password | b64enc) | b64enc }}
{{- end }}

{{- define "stackrox-io-pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .Values.pullSecrets.stackrox.registry (printf "%s:%s" .Values.pullSecrets.stackrox.username .Values.pullSecrets.stackrox.password | b64enc) | b64enc }}
{{- end }}


{{- define "pull-secret" }}
  {{- printf "{\"auths\": {\"%s\": {\"auth\": \"%s\"}}}" .registry (printf "%s:%s" .username .password | b64enc) | b64enc }}
{{- end }}