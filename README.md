# Pirsch Go Proxy

A self-hosted proxy for the Pirsch Analytics JavaScript snippet.

## Why should I use a proxy?

The benefit of using a proxy is that your website will only make first-party requests. The JavaScript snippets are hosted on your own server. Requests to pirsch.io will be proxied through your server, preventing them from being blocked by ad blockers.

Additionally, you can create rollup views and send data to multiple dashboards with a single request on the client.

## Installation

Download the latest release archive from the release section on GitHub and extract it onto your server. Adjust the pirsch/config.toml file to your needs.

```toml
# Proxy server configuration.
# You should use a TLS certificate or run it behind a reverse proxy that queries a certificate for you.
[server]
    host = ":80"
    write_timeout = 5
    read_timeout = 5
    #tls = true
    #tls_cert = "path/to/cert_file"
    #tls_key = "path/to/key_file

# Proxy network configuration.
# This configuration can be used to retreive the real client IP address and set accepted subnets for proxies or load balancers in front of this proxy.
[network]
    # Parsed in order, allowed values: CF-Connecting-IP, True-Client-IP, X-Forwarded-For, Forwarded, X-Real-IP
    # The proxy will use the remote IP if no header is configured.
    header = ["X-Forwarded-For", "Forwarded"]

    # List of CIDR.
    subnets = ["10.0.0.0/8"]

# List of clients to send data to.
# The client ID can be left empty if you use an access token instead of oAuth.
[[clients]]
    id = "your-client-id"
    secret = "your-client-secret"
    hostname = "example.com"

#[[clients]]
#    id = "your-client-id"
#    secret = "your-client-secret"
#    hostname = "example.com"

# ...
```

`clients` takes a list of API clients. You can create a new client ID and secret on the Pirsch dashboard on the developer settings page. The hostname needs to match the hostname you have configured on the dashboard.

The proxy will send all page views and events to all clients configured. So, if you would like to send the statistics to two dashboards, you can add another client by appending it to the list.

## Docker

Alternatively you can use Docker to install the proxy. Here is a docker-compose to deploy it.

```yaml
version: "3"

services:
  pirsch-proxy:
    image: pirsch/proxy
    container_name: pirsch-proxy
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config.toml:/app/config.toml
```

## Usage

After you have installed the proxy on your server, you can add the Pirsch JavaScript snippet to your website.

**pirsch.min.js**

This will track page views.

```JavaScript
<script defer type="text/javascript" src="/pirsch/pirsch.min.js" id="pirschjs"></script>
```

**pirsch-events.min.js**

This will make the `pirsch` event function available on your site.

```JavaScript
<script defer type="text/javascript" src="/pirsch/pirsch-events.min.js" id="pirscheventsjs"></script>
```

If you have installed it on a different domain or subdomain, adjust `src` and the endpoints using the `data-endpoint` parameters.

```JavaScript
<script defer type="text/javascript"
    src="https://tracking.example.com/pirsch/pirsch.min.js"
    id="pirschjs"
    data-endpoint="https://tracking.example.com/pirsch/hit"></script>

<script defer type="text/javascript"
    src="https://tracking.example.com/pirsch/pirsch-events.min.js"
    id="pirscheventsjs"
    data-endpoint="https://tracking.example.com/pirsch/event"></script>
```

A demo can be found in the [demo](demo) directory.

## Local development

The `config.toml` takes a `base_url` parameter to configure a local Pirsch mock implementation.

```toml
base_url = "http://localhost:8080"
```

## License

MIT
