{{- define "k.cronjob" -}}
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ .Release.Name }}-{{ .Release.Revision }}-cronjob
  annotations:
    iter8.tools/group: {{ .Release.Name }}
    iter8.tools/revision: {{ .Release.Revision | quote }}
spec:
  schedule: {{ .Values.cronjobSchedule | quote }}
  concurrencyPolicy: Forbid
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            iter8.tools/group: {{ .Release.Name }}
          annotations:
            sidecar.istio.io/inject: "false"
        spec:
          serviceAccountName: {{ default (printf "%s-iter8-sa" .Release.Name) .Values.serviceAccountName }}
          containers:
          - name: iter8
            image: {{ .Values.iter8Image }}
            imagePullPolicy: Always
            command:
            - "/bin/sh"
            - "-c"
            - |
              iter8 k run --namespace {{ .Release.Namespace }} --group {{ .Release.Name }} -l {{ .Values.logLevel }} --reuseResult
            resources:
              {{ toYaml .Values.resources | indent 14 | trim }}
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsNonRoot: true
              runAsUser: 1001040000
          restartPolicy: Never
      backoffLimit: 0
{{- end }}
