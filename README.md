# Webhooked

A webhook receiver on steroids. The process is simple, receive webhook from all over the world, and send it to your favorite pub/sub to process it immediately or later without losing any received data 

![Webhooked explained](/.github/profile/webhooked.png)

## Motivation

When you start working with webhooks, it's often quite random, and sometimes what shouldn't happen, does. **One or more data sent by a webhook is lost because our service did not respond, or worse to crash**. That's why very often it's better to make a small HTTP server that only receives and conveys the information to another service that will process the information.

This is exactly what `Webhooked` does !

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
  # value. If the header value is equals to `test`, `foo` or `bar`, the webhook is
  # process. Else no process is handled and http server return a 401 error
  security:
  - getHeader:
      name: X-Hook-Secret
  - compareWithStaticValue:
      value: 'test'
      values: ['foo', 'bar']
```

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
