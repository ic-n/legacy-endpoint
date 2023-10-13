# Legacy-Endpoint Go Package

## Introduction

The `legend` (<b>Leg</b>acy-<b>end</b>point) package is a Go library that provides a solution for gradually replacing significant monolithic endpoints with a more lightweight client-oriented approach. This package is designed to help transition from heavy monolithic endpoints to a more modular architecture without breaking compatibility with old clients. The package facilitates this transition by allowing you to reroute and decouple massive endpoints step by step using a fragmented JSON response joined by `legend`. By implementing this package, you can make incremental changes and improvements to your API while supporting existing clients.

The most challenging part with monolytic endpoints is to cache some specific fields. The `legend` package provides built-in caching functionality using the [bigcache](https://github.com/allegro/bigcache) library. Caching is optional and can be configured when adding collectors. If caching is enabled, the data will be cached to reduce the load on your source.

## Example

For example, we had some static data in the database varying by version and some URL managing microservice that returned image URLs showing the flowers of the day.

Check it to understand usecase [examples/static-media/](examples/static-media/media/handler_test.go)

## Concerns

_This package is an experiment, you can use it or be inspired_

The `legend` package is designed to work alongside existing endpoints, making it easier to transition to a more modular and client-friendly architecture while continuing to serve your old clients.

By implementing the `legend` package, you can gradually transform your API architecture to be more modular, efficient, and respectful of old and new clients. This transition enables you to move forward without disrupting existing client functionality.

Feel free to customize and extend the package to suit your requirements and legacy API endpoints.

