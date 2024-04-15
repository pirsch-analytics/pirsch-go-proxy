# Pirsch Go Proxy

A self-hosted proxy for the [Pirsch Analytics](https://pirsch.io) JavaScript snippets.

## Why should I use a proxy?

The benefit of using a proxy is that your website will only make first-party requests. The JavaScript snippets are hosted on your own server. Requests to pirsch.io will be proxyed through your server, preventing them from being blocked by ad blockers.

Additionally, you can create rollup views and send data to multiple dashboards with a single request from the client.

## Installation

Download the latest release archive from the releases section on GitHub and extract it to your server. Create an API client (or several) on the Pirsch dashboard and edit the [config.toml](config/config.toml) file to suit your needs. Then you can start the server. We recommend creating a systemd unit file or using Docker. The configuration path can be passed as the first application argument.

## Docker

Alternatively, you can use Docker to install the proxy. A docker-compose for deployment can be found [here](deploy/docker-compose.yml);

## Usage

Once you have installed the proxy on your server, you can add the Pirsch JavaScript snippet to your website.

> If you have adjusted the path configuration, make sure you adjust the script and endpoint paths.
> If you have installed it on a different domain or subdomain, adjust the `src` and `data-endpoint` parameters to include the domain.

Here is an example with the default configuration:

```JavaScript
<script defer type="text/javascript"
    src="/p/pa.js"
    id="pianjs"
    data-hit-endpoint="/p/pv"
    data-event-endpoint="/p/e"
    data-session-endpoint="/p/s"></script>
```

## Local development

The `config.toml` takes a `base_url` parameter to configure a local Pirsch mock implementation.

## License

MIT
