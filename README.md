# Search Backend for Hugo with ElasticSearch

Requirements:
* For generating indexes:
    ```
        npm install -g kovetskiy/hugo-elasticsearch
    ```
* ElasticSearch 

How does it work:
1. Generates index using hugo-elasticsearch which just reads all markdown
   files and generates index that can be uploaded into elasticsearch
2. Uploads that file and adds new alias
3. Removes old aliases and old indexes

Steps 2 and 3 used just to avoid broken search engine.

Environment variables:
* `LISTEN` - address to listen
* `ELASTIC` - address of elastic
* `INPUT` - path with content like that: "/srv/hugo/content/**"
* `LANGUAGE` - yaml or toml
* `DELIMITER` - `---` for yaml or `+++` for toml
* `INDEX` - name of index, I prefer to call it the same as website
