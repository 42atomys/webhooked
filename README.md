<h1 align="center">Webhooked</h1>

<p align="center"><a href="https://github.com/42Atomys/webhooked/actions/workflows/release.yaml"><img src="https://github.com/42Atomys/webhooked/actions/workflows/release.yaml/badge.svg" alt="Release 🎉"></a>
<a href="https://goreportcard.com/report/atomys.codes/webhooked"><img src="https://goreportcard.com/badge/atomys.codes/webhooked" /></a>
<a href="https://codeclimate.com/github/42Atomys/webhooked"><img alt="Code Climate maintainability" src="https://img.shields.io/codeclimate/maintainability/42Atomys/webhooked"></a>
<a href="https://codecov.io/gh/42Atomys/webhooked"><img alt="Codecov" src="https://img.shields.io/codecov/c/gh/42Atomys/webhooked?token=NSUZMDT9M9"></a>
<img src="https://img.shields.io/github/v/release/42atomys/webhooked?label=last%20release" alt="GitHub release (latest by date)">
<img src="https://img.shields.io/github/contributors/42Atomys/webhooked?color=blueviolet" alt="GitHub contributors">
<img src="https://img.shields.io/github/stars/42atomys/webhooked?color=blueviolet" alt="GitHub Repo stars">
<a href="https://hub.docker.com/r/atomys/webhooked"><img src="https://img.shields.io/docker/pulls/atomys/webhooked" alt="Docker Pull"></a>
<a href="https://hub.docker.com/r/atomys/webhooked"><img src="https://img.shields.io/docker/image-size/atomys/webhooked" alt="Docker Pull"></a>
<a href="https://pkg.go.dev/atomys.codes/webhooked"><img src="https://pkg.go.dev/badge/atomys.codes/webhooked.svg" alt="Go Reference"></a></p>
  
<p align="center">A webhook receiver on steroids. The process is simple, receive webhook from all over the world, and send it to your favorite pub/sub to process it immediately or later without losing any received data</p>

<p align="center"><img src="/.github/profile/webhooked.png" alt="Webhooked explained"></p>

## Motivation

When you start working with webhooks, it's often quite random, and sometimes what shouldn't happen, does. **One or more data sent by a webhook is lost because our service did not respond, or worse to crash**. That's why very often it's better to make a small HTTP server that only receives and conveys the information to another service that will process the information.

This is exactly what `Webhooked` does !

## Roadmap

I am actively working on this project in order to release a stable version for **2025**

![Roadmap](/.github/profile/roadmap.png)

## Usage

### Step 1 : Configuration file
```yaml
apiVersion: v1alpha1
# List of specifications of your webhooks listerners.
specs:
- # Name of your listener. Used to store relative datas and printed on log
  name: exampleHook
  # The Entrypoint used to receive this Webhook
  # In this example the final url will be: example.com/v1alpha1/webhooks/example
  entrypointUrl: /webhooks/example
  # Security factories used to verify the payload 
  # Factories is powerful and very modular. This is executed in order of declaration
  # and need to be ended by a `compare` Factory.
  #
  # In this example we get the header `X-Hook-Secret` and compare it to a static
  # value. If the header value is equals to `test`, `foo` or `bar`, or the value
  # contained in SECRET_TOKEN env variable, the webhook is process. 
  # Else no process is handled and http server return a 401 error
  #
  # If you want to use insecure (not recommended), just remove security property
  security:
  - header:
      inputs:
      - name: headerName
        value: X-Hook-Secret
  - compare:
      inputs:
      - name: first
        value: '{{ Outputs.header.value }}'
      - name: second
        values: ['foo', 'bar']
        valueFrom:
          envRef: SECRET_TOKEN
  
  # Formatting allows you to apply a custom format to the payload received
  # before send it to the storage. You can use built-in helper function to
  # format it as you want. (Optional)
  #
  # Per default the format applied is: "{{ .Payload }}"
  # 
  # THIS IS AN ADVANCED FEATURE :
  # Be careful when using this feature, the slightest error in format can
  # result in DEFFINITIVE loss of the collected data. Make sure your template is
  # correct before applying it in production.
  formatting:
    templateString: |
      {
        "config": "{{ toJson .Config }}",
        "metadata": {
          "specName": "{{ .Spec.Name }}",
          "deliveryID": "{{ .Request.Header | getHeader "X-Delivery" | default "unknown" }}"
        },
        "payload": {{ .Payload }}
      }

  # Storage allows you to list where you want to store the raw payloads
  # received by webhooked. You can add an unlimited number of storages, webhooked
  # will store in **ALL** the listed storages
  # 
  # In this example we use the redis pub/sub storage and store the JSON payload
  # on the `example-webhook` Redis Key on the Database 0
  storage:
  - type: redis
    # You can apply a specific formatting per storage (Optional)
    formatting: {}
    # Storage specification
    specs:
      host: redis.default.svc.cluster.local
      port: 6379
      database: 0
      key: example-webhook


  # Response is the final step of the pipeline. It allows you to send a response
  # to the webhook sender. You can use the built-in helper function to format it
  # as you want. (Optional)
  #
  # In this example we send a JSON response with a 200 HTTP code and a custom
  # content type header `application/json`. The response contains the deliveryID
  # header value or `unknown` if not present in the request.
  response:
    formatting:
      templateString: |
        {
          "deliveryID": "{{ .Request.Header | getHeader "X-Delivery" | default "unknown" }}"
        }
    httpCode: 200
    contentType: application/json
```

More informations about security pipeline available on wiki : [Configuration/Security](https://github.com/42Atomys/webhooked/wiki/Security)

More informations about storages available on wiki : [Configuration/Storages](https://github.com/42Atomys/webhooked/wiki/Storages)

More informations about formatting available on wiki : [Configuration/Formatting](https://github.com/42Atomys/webhooked/wiki/Formatting)

### Step 2 : Launch it 🚀
### With Kubernetes

If you want to use kubernetes, for production or personnal use, refere to example/kubernetes:

https://github.com/42Atomys/webhooked/tree/main/examples/kubernetes


### With Docker image

You can use the docker image [atomys/webhooked](https://hub.docker.com/r/atomys/webhooked) in a very simplistic way

```sh
# Basic launch instruction using the default configuration path
docker run -it --rm -p 8080:8080 -v ${PWD}/myconfig.yaml:/config/webhooked.yaml atomys/webhooked:latest
# Use custom configuration file
docker run -it --rm -p 8080:8080 -v ${PWD}/myconfig.yaml:/myconfig.yaml atomys/webhooked:latest serve --config /myconfig.yaml
```

### With pre-builded binary

```sh
./webhooked serve --config config.yaml -p 8080
```

## To-Do

TO-Do is moving on Project Section: https://github.com/42Atomys/webhooked/projects?type=beta

# Contribution

All pull requests and issues on GitHub will welcome.

All contributions are welcome :)

## Thanks
