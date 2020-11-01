# racce
Race results for Assetto Corsa Competizione

# Installation
1. Clone repository into desired installation location
2. Copy `configuration.default.json` to `configuration.json`
3. Change configuration as appropriate (see below)
4. Execute `go run .` to launch the server

# Configuration
The configuration is read during startup from a JSON file called `configuration.json`. It contains the following top-level objects:

| Name     | Description                                        |
|----------|----------------------------------------------------|
| frontend | Configuration for the web frontend                 |
| server   | Configuration for the accServer that is being used |

## Frontend

The frontend can be configured with the following settings:

| Name          | Required | Description                                                                                     |
|---------------|----------|-------------------------------------------------------------------------------------------------|
| listen        | yes      | IP and port to listen on in the format "ip:port". Leave the IP out to listen on all interfaces. |
| adminPassword | no       | The password required to access the admin pages. Leave empty to disable the admin pages.        |

## Server

The accServer that is being used can be used for the results only, or it can also be managed via the admin pages. The settings indicate how the accServer is being used.

| Name            | Required | Description |
|-----------------|----------|-------------------------------------------------------------------------------------------------|
| installationDir | no*      | The path to the acc server directory where the accServer is installed. The path must contain forwarded slashes, even on Windows. If the `installationDir` is present and contains a valid accServer, the server can be managed via the admin pages. |
| resultsDir      | no*      | The path where the JSON results files are stored by the accServer. This defaults to the `results/` subdirectory of the `installationDir` if not given. |
| newResultsDelay | yes      | Number of seconds to wait after a new results file was written before it is read. This should not be 0 to avoid reading files which are still being written. The default of 5 should be fine in nearly all circumstances. |

(*) At least one of `installationDir` or `resultsDir` must be specified.

# HTTP forwarding

The HTTP server in racce is a basic application server and support for more advanced features like SSL are not exposed. You can use a more complete HTTP server, such as nginx, to handle these and forward the requests to the racce webserver.

To allow racce to generate proper links when forwarded from a subdirectory, specify the `X-Forwarded-Prefix` header. Websocket connections also need to be forwarded correctly to allow live monitoring of the server log. Consider the following example for nginx, which forwards the subdirectory `/acc` to a racce server running on localhost:

    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }
    location /acc/ {
        proxy_pass http://localhost:8099/;
        proxy_http_version 1.1;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-Prefix /acc;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
    }
