# michiru
_Named after the ["girl on the right"](https://wiki.anidb.net/Who_is_that_girl)_

A minimal webserver in Go, providing out-of-the-box title searching of anime titles from AniDB, backed by Meilisearch.

The primary function of this application is to make it easier to obtain the `aid` of any given anime, which will then allow users to work with other parts of the AniDB API.

## Docker Compose Setup

Easily self-hostable by simply copying the latest `docker-compose.yml` from this repository:

```yaml
services:
  server:
    image: ghcr.io/chaaaaun/michiru-server:latest
    ports:
      - "127.0.0.1:8080:8080"
    env_file:
      - ".env"
    volumes:
      - ./static:/static
    restart: unless-stopped
    depends_on:
      - meilisearch
  importer:
    image: ghcr.io/chaaaaun/michiru-importer:latest
    env_file:
      - ".env"
    depends_on:
      - meilisearch
  meilisearch:
    image: getmeili/meilisearch:v1.14
    env_file:
      - ".env"
    volumes:
      - ./meili_data:/meili_data
    restart: unless-stopped
    command: ["meilisearch", "--no-analytics"]
```

Then copy the `.env.template` into the same directory as the `docker-compose.yml` and rename as `.env`:

```conf
#####
# Env vars for the meilisearch container
#####
# Refer to meilisearch documentation for more details
MEILI_ENV=development
MEILI_MASTER_KEY=

#####
# Env vars for the importer container
#####
# Should match MEILI_MASTER_KEY 
MEILISEARCH_KEY=

# Should correspond to the port which the meilisearch instance is listening on,
# if you've changed the config for that
MEILISEARCH_URL=http://meilisearch:7700

# Meilisearch index name under which title data is stored, defaults to "titles"
# INDEX_NAME=

# How long before the importer times out a meilisearch job, defaults to 0
# TASK_TIMEOUT=

# URL to retrieve the compressed XML title dump from AniDB
TITLE_DUMP_URL=https://anidb.net/api/anime-titles.xml.gz
# How long before the importer times out a fetch request to above, defaults to 30s
# FETCH_TIMEOUT=


#####
# Env vars for the server container
#####
# Port the server listens on, defaults to 8080
# PORT=

# Relative path for where the server looks to serve a frontend,
# optional if you just want to host the API
WEBUI_PATH=./static
```

Generate a secure API key using your preferred method and populate both `MEILI_MASTER_KEY` and `MEILISEARCH_KEY` with that key.

Simply run `docker compose up -d`.

## Building and Running

```shell
git clone https://github.com/chaaaaun/michiru.git
cd michiru
CGO_ENABLED=1 go build cmd/importer
CGO_ENABLED=0 go build cmd/server
```

The same environment variables documented above should be provided before running the built binaries.
The importer requires the `libxml2` package to be installed before building.

## API Reference

`/search`
> Search for AniDB AID using fuzzy multilingual title search
 
**Query Parameters**

| Parameter | Type    | Required            | Description                                            |
|-----------|---------|---------------------|--------------------------------------------------------|
| `query`   | string  | true                | The full or partial name of the anime you want to find |
| `offset`  | integer | false (default: 0)  | How many results to offset before returning            |
| `limit`   | integer | false (default: 10) | How many results to return in the response             |

<details>
<summary>Example response for <code>/search?query=test&limit=1</code></summary>

```json
{
    "payload": [
        {
            "aid": 357,
            "mainTitle": "Test Anime",
            "officialTitles": {
                "ja": [
                    "ンート"
                ]
            },
            "synonymousTitles": {
                "x-jat": [
                    "test`blubb"
                ]
            },
            "kanaTitles": {
                "ja": [
                    "test title k"
                ]
            },
            "cardTitles": {
                "ja": [
                    "test title c"
                ],
                "x-jat": [
                    "test title c TR"
                ]
            },
            "_formatted": {
                "aid": 357,
                "mainTitle": "<span>Test</span> Anime",
                "officialTitles": {
                    "ja": [
                        "ンート"
                    ]
                },
                "synonymousTitles": {
                    "x-jat": [
                        "<span>test</span>`blubb"
                    ]
                },
                "kanaTitles": {
                    "ja": [
                        "<span>test</span> title k"
                    ]
                },
                "cardTitles": {
                    "ja": [
                        "<span>test</span> title c"
                    ],
                    "x-jat": [
                        "<span>test</span> title c TR"
                    ]
                }
            },
            "_rankingScore": 0.6666666666666666
        }
    ],
    "paging": {
        "count": 19,
        "next": "/search?limit=1&offset=1&query=test"
    }
}
```

</details>

`/metadata`
> Metadata about the latest-retrieved title dump and meilisearch index

<details>
<summary>Example response for <code>/metadata</code></summary>

```json
{
    "id": "titles",
    "retrievedAt": "2025-07-27T02:00:02Z",
    "updatedAt": "2025-07-26T03:00:07Z",
    "dumpEntries": 16172,
    "dumpTitles": 95683
}
```

</details>
