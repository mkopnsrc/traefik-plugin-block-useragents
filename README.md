# Traefik Plugin: Block User-Agent

A Traefik middleware plugin to block all HTTP requests by default and allowing only HTTP requests based on the user-specified `User-Agent` patterns for Browser and OS types.

## Features
- Blocks all `User-Agent` by default.
- Allows user-defined browsers via:
  - Custom regex patterns (e.g., `MyBrowser/12[0-1].*`).
- Optional OS type filtering with regex patterns. See example below.
- No external APIs or caching; relies entirely on user configuration.

## Notes
 - Requirements: At least one `allowedBrowsers` entry with `name` and it's `regex` is required.
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
              regex: "Chrome/13[0-3].*" # Chrome 130-133
            - name: "Firefox"
              regex: "Firefox/13[1-5].*" # Firefox 131-135
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
              regex: "Chrome/13[0-3].*" # Chrome 130-133
            - name: "Firefox"
              regex: "Firefox/13[1-5].*" # Firefox 131-135
            - name: "Edg" # Microsoft Edge
              regex: "Edg/10[0-9]" # Edge 100-199
            - name: "Brave" 
              regex: "Brave/1.[7][5-9]" # Brave 1.75-1.99
            - name: "CriOS" # Chrome for iOS
              regex: "CriOS/13[0-9]"
          allowedOSTypes:
            - "Windows NT 10\\.0" # Windows 10
            - "Mac OS X 10\\.[0-9]+" # macOS 10.x
            - "Linux" # Linux
            - "X11" # Unix
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
