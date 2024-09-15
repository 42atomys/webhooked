{
  "config": "{{ toJson .Config }}",
  "storage": {{ toJson .Storage }},
  "metadata": {
    "model": "{{ .Request.Header.Get "X-Model" }}",
    "event": "{{ .Request.Header.Get "X-Event" }}",
    "deliveryID": "{{ .Request.Header.Get "X-Delivery" | default "unknown" }}"
  },
  "payload": {{ .Payload }}
}
