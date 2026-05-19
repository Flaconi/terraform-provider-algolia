# Terraform Provider Algolia

[![License: MPL-2.0](https://img.shields.io/badge/License-MPL2.0-blue.svg)](./LICENSE)
[![Tests Workflow](https://github.com/Flaconi/terraform-provider-algolia/workflows/Tests/badge.svg)](https://github.com/Flaconi/terraform-provider-algolia/actions/workflows/test.yml)

Terraform Provider for [Algolia](https://www.algolia.com).

> **Fork notice.** This is the [Flaconi/terraform-provider-algolia](https://github.com/Flaconi/terraform-provider-algolia) fork of [k-yomo/terraform-provider-algolia](https://github.com/k-yomo/terraform-provider-algolia), upgraded to use [`algoliasearch-client-go` **v4**](https://github.com/algolia/algoliasearch-client-go) (the upstream main branch is still on v3). See [Fork differences](#fork-differences) below.

## Documentation

Full, comprehensive documentation for the underlying provider schema is available on the Terraform website:

[https://registry.terraform.io/providers/k-yomo/algolia/latest/docs](https://registry.terraform.io/providers/k-yomo/algolia/latest/docs)

The schema in this fork is API-compatible with upstream; the Algolia SDK upgrade is internal.

## Using the provider

The provider authenticates against Algolia using two values:

| Env var | Description |
|---|---|
| `ALGOLIA_APP_ID` | Your Algolia application ID (10-character string from the dashboard). Required. Can also be set via the `app_id` provider attribute. |
| `ALGOLIA_API_KEY` | An Algolia **Admin API Key**. Required. A search-only or scoped key will fail on most resources. Can also be set via the `api_key` provider attribute. |

```sh
$ export ALGOLIA_APP_ID=<your app id>
$ export ALGOLIA_API_KEY=<your admin api key>
```

The example below demonstrates the following operations:

- create index
- create rule for the index
- create api key to search the index

```terraform
terraform {
  required_providers {
    algolia = {
      source  = "Flaconi/algolia"
      version = ">= 0.1.0, < 1.0.0"
    }
  }
}

provider "algolia" {
  app_id = "XXXXXXXXXX"
}

resource "algolia_index" "example" {
  name = "example"
  attributes_config {
    searchable_attributes = [
      "title",
      "category,tag",
      "unordered(description)",
    ]
    attributes_for_faceting = [
      "category"
    ]
    unretrievable_attributes = [
      "author_email"
    ]
    attributes_to_retrieve = [
      "title",
      "category",
      "tag",
      "description",
      "body"
    ]
  }

  ranking_config {
    ranking = [
      "words",
      "proximity"
    ]
  }

  faceting_config {
    max_values_per_facet = 50
    sort_facet_values_by = "alpha"
  }

  languages_config {
    remove_stop_words_for = ["en"]
  }
}

resource "algolia_rule" "example" {
  index_name = algolia_index.example.name
  object_id  = "example-rule"

  conditions {
    pattern   = "{facet:category}"
    anchoring = "contains"
  }

  consequence {
    params_json = jsonencode({
      automaticFacetFilters = [
        {
          facet       = "category"
          disjunctive = true
          score       = 0
        }
      ]
    })
  }
}


resource "algolia_api_key" "example" {
  acl = [
    "search",
    "browse"
  ]
  expires_at = "2030-01-01T00:00:00Z"
  max_hits_per_query = 100
  max_queries_per_ip_per_hour = 10000
  description = "This is a example api key"
  indexes = [algolia_index.example.name]
  referers = ["https://algolia.com/\\*"]
}
```

## Supported resources

- [x] [Settings](https://www.algolia.com/doc/api-client/methods/settings/)
- [x] [Rule](https://www.algolia.com/doc/api-client/methods/rules/)
- [x] [Api Keys](https://www.algolia.com/doc/api-client/methods/api-keys/)
- [x] [Synonym](https://www.algolia.com/doc/api-client/methods/synonyms/)
- [x] [Query Suggestions](https://www.algolia.com/doc/rest-api/query-suggestions/)
- [ ] [A/B Test](https://www.algolia.com/doc/api-client/methods/ab-test/)
- [ ] [Dictionaries](https://www.algolia.com/doc/api-client/methods/dictionaries/)
- [ ] [Personalization](https://www.algolia.com/doc/api-client/methods/personalization/)

## Fork differences

This fork tracks [k-yomo/terraform-provider-algolia](https://github.com/k-yomo/terraform-provider-algolia) but adds:

- **Algolia Go client upgraded to v4** (`algoliasearch-client-go/v4`). Upstream main is still on v3; upstream's [PR #269](https://github.com/k-yomo/terraform-provider-algolia/pull/269) attempts the v4 upgrade but has failing CI and has not been merged.
- **Read normalization for v4's response shape.** The v4 SDK returns `nil` for fields at server-side defaults, which would otherwise produce permanent plan diffs. The provider now substitutes the documented defaults (`["*"]` for `attributes_to_retrieve`, `"…"` for `snippet_ellipsis_text`, `100` for `relevancy_strictness`, `10` for `max_facet_hits`, etc.).
- **Last-replica detach fix.** When deleting the only replica of a primary, the old code passed a `nil` slice to `SetSettings`, which v4's `omitempty`-aware `MarshalJSON` then omitted entirely — so the replicas list was never updated and `DeleteIndex` failed with 403. Fixed to always send an explicit empty array.
- **Replica detach propagation poll.** After clearing the primary's replicas list, the provider now polls the replica's `primary` setting until the detachment has propagated server-side before issuing `DeleteIndex`.
- **API key Create/Update wait workarounds.** The v4 SDK's `WaitForApiKey(ADD)` treats 403 propagation responses as success (premature exit), and `WaitForApiKey(UPDATE)` compares the ticking-down `Validity` field (never matches). The provider replaces both with its own polling loop.
- **Rule condition `context` empty-string fix.** v4's API rejects empty `context` values against the regex `[A-Za-z0-9_-]+`. The provider no longer sends a pointer-to-empty-string when the field is unset.
- **Query suggestions idempotent delete.** Algolia auto-deletes the query-suggestions config when its source index is deleted, so `DeleteConfig` returning 404 is now treated as success.
- **Unified `IsNotFoundError`** across the search and query-suggestions client error types.

## Contributing

I appreciate your help!

To contribute, please read the [Contributing to Terraform - Algolia Provider](./CONTRIBUTING.md)

## Sponsor
<a href="https://www.manomano.com">
  <img src="https://avatars.githubusercontent.com/u/25135828?s=200&v=4" width="64">
  <br>
  ManoMano
</a>
