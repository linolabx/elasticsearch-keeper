# Elasticsearch Keeper

## Usage

```bash
docker run --rm -it -p 9000:80 ghcr.io/linolabx/elasticsearch-keeper:latest \
  -v <path-to-synonyms-config-dir>:/synonyms \
  -e API_KEY=<api-key> \
  -e SYNONYMS_CONFIG_DIR=/synonyms \
  -e REDIS_URL=<redis-url> \
  -e REDIS_PREFIX=<redis-prefix>

curl -X POST "localhost:9000/synonyms/<filename>" -H 'Authorization: Bearer <api-key>' \
 -F 'file=@<path-to-synonyms-file>' \
 -F 'indexes[]=<index-name-1>' \
 -F 'indexes[]=<index-name-2>'
```
