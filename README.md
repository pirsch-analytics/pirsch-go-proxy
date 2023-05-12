# Pirsch Go Proxy

A self-hosted proxy for the Pirsch Analytics JavaScript snippets.

## Why should I use a proxy?

The benefit of using a proxy is that your website will only make first-party requests. The JavaScript snippets are hosted on your own server. Requests to pirsch.io will be proxyed through your server, preventing them from being blocked by ad blockers.

Additionally, you can create rollup views and send data to multiple dashboards with a single request from the client.

## Installation

Download the latest release archive from the releases section on GitHub and extract it to your server. Create an API client (or several) on the Pirsch dashboard and edit the [config.toml](config.toml) file to suit your needs. You can then start the server. We recommend creating a systemd unit file or using Docker.

## Docker

Alternatively, you can use Docker to install the proxy. A docker-compose for deployment can be found [here](docker-compose.yml);

## Usage

Once you have installed the proxy on your server, you can add the Pirsch JavaScript snippet to your website.

> If you have adjusted the path configuration, make sure you adjust the script and endpoint paths.
> If you have installed it on a different domain or subdomain, adjust the `src` and `data-endpoint` parameters to include the domain.

Here is an example for the `pirsch.js` script with the default configuration.

```JavaScript
<script defer type="text/javascript"
        src="/p/p.js"
        id="pirschjs"
        data-endpoint="/p/pv"></script>
```

There are three other scripts available:

* `pirsch-events.js` as `e.js` using the endpoint `/p/e`
* `pirsch-sessions.js` as `s.js` using the endpoint `/p/s`
* `pirsch-extended.js` as `ext.js` using all of the other endpoints

Note that the extended scripts use different endpoint parameters. Namely `data-hit-endpoint`, `data-event-endpoint` and `data-session-endpoint`.

## Local development

The `config.toml` takes a `base_url` parameter to configure a local Pirsch mock implementation.

## License

MIT
