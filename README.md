# Webhooked

A webhook receiver on steroids. The process is simple, receive webhook from all over the world, and send it to your favorite pub/sub to process it immediately or later without losing any received data 

![Webhooked explained](/.github/profile/webhooked.png)

## Motivation

When you start working with webhooks, it's often quite random, and sometimes what shouldn't happen, does. **One or more data sent by a webhook is lost because our service did not respond, or worse to crash**. That's why very often it's better to make a small HTTP server that only receives and conveys the information to another service that will process the information.

This is exactly what `Webhooked` does !

## Roadmap

I am actively working on this project to release a stable version by the **end of March 2022**

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
  # Storage allows you to list where you want to store the raw payloads
  # received by webhooked. You can add an unlimited number of storages, webhooked
  # will store in **ALL** the listed storages
  # 
  # In this example we use the redis pub/sub storage and store the JSON payload
  # on the `example-webhook` Redis Key on the Database 0
  storage:
  - type: redis
    specs:
      host: redis.default.svc.cluster.local
      port: 6379
      database: 0
      key: example-webhook
```

More informations about security pipeline available on wiki : [Configuration/Security](https://github.com/42Atomys/webhooked/wiki/Security)

More informations about storages available on wiki : [Configuration/Storages](https://github.com/42Atomys/webhooked/wiki/Configuration-Storages)

### Step 2 : Launch it 🚀
### With Kubernetes

If you want to use kubernetes, for production or personnal use, refere to example/kubernetes:

https://github.com/42Atomys/webhooked/tree/main/examples/kubernetes


### With Docker image

You can use the docker image [atomys/webhooked](https://hub.docker.com/r/atomys/webhooked) in a very simplistic way

```sh
# Basic launch instruction using the default configuration path
docker run -it --rm -p 8080:8080 -v ${PWD}/myconfig.yaml:/config/webhooks.yaml atomys/webhooked:latest
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
