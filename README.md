# Traefik Plugin: Block User-Agent

A Traefik middleware plugin to block HTTP requests based on the `User-Agent` header, allowing only user-specified browser patterns.

## Features
- Blocks all `User-Agent`s by default.
- Allows user-defined browsers via:
  - Custom regex patterns (e.g., `MyBrowser/12[0-1].*`).
  - Version thresholds (e.g., `>121` generates `MyBrowser/121.*`).
- Optional OS type filtering with regex patterns.
- No external APIs or caching; relies entirely on user configuration.
- Browser names are dynamic and used as provided in regex generation.

## Notes
 - Requirements: At least one `allowedBrowsers` entry with either `regex` or `version` is required. If none are provided, all requests will be blocked.
 - Dynamic Browser Names: The `name` field in `allowedBrowsers` is used as-is in regex generation when `version` is provided (e.g., `CustomBot/1.0.*`).
 - Version Handling: The `version` field with > generates a regex (e.g., `>121` becomes `Browser/121.*`). Without `>`, it matches exactly (e.g., `121` becomes `Browser/121.*`).
 - OS Patterns: `allowedOSTypes` expects regex patterns. Use exact strings (e.g., `Windows NT 10\.0`) or wildcards (e.g., `Android [8-9]\.[0-9]+`) as needed.
 - No Dependencies: The plugin is lightweight with no external dependencies.



## Usage
1. Add the plugin to your Traefik configuration.
2. Configure the plugin with the desired browser patterns.
3. Attach the plugin to your Traefik middleware.

## Traefik Experimental Plugin Registry (traefik.yml)
```yaml
experimental:
  plugins:
    block_useragents:
      moduleName: "github.com/mkopnsrc/traefik-plugin-block-useragents"
      version: "v1.0" # Optional
```

## Traefik Local Plugin (traefik.yml)
### Ensure Local Plugin directory is mounted in the Traefik container.
```yaml
experimental:
  localPlugins:
    block_useragents:
      moduleName: "github.com/mkopnsrc/traefik-plugin-block-useragents"
```

## Middleware Configuration
### Browsers Only
```yaml
http:
  middlewares:
    block-ua:
      plugin:
        block_useragents:
          allowedBrowsers:
            - name: "Chrome"
              regex: "Chrome/12[0-1].*"
            - name: "Firefox"
              version: ">122"
```


### Browsers with OS Types Filtering
```yaml
http:
  middlewares:
    block-ua:
      plugin:
        block_useragents:
          allowedBrowsers:
            - name: "Chrome"
              version: ">130"
            - name: "Firefox"
              regex: "Firefox/13[1-5].*"
            - name: "Safari"
              version: ">533"
            - name: "Edg" # Microsoft Edge
              version: ">100"
            - name: "Brave" 
              version: ">1.75"
            - name: "CriOS" # Chrome for iOS
              regex: "CriOS/13[0-9]"
          allowedOSTypes:
            - "Windows NT 10\.0" # Windows 10
            - "Mac OS X 10\.[0-9]+" # macOS 10.x
            - "Linux" # Linux
            - "Android" # Android
            - "iOS" # iOS
```

## Router Usage
```yaml
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service
      middlewares:
        - block-ua
```