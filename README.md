# michiru
_Named after the ["girl on the right"](https://wiki.anidb.net/Who_is_that_girl)_

A minimal webserver in Go, providing out-of-the-box title searching of anime titles from AniDB, backed by Meilisearch.
The primary function of this application is to make it easier to obtain the `aid` of any given anime, which will then allow users to work with other parts of the AniDB API.

## Installation
`go build cmd/main.go`
