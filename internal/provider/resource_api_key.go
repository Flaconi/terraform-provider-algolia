package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-algolia/internal/algoliautil"
)

func resourceAPIKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIKeyStateContext,
		},
		Description: "A configuration for an API key",
		// https://www.algolia.com/doc/api-reference/api-methods/add-api-key/
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The created key.",
			},
			"acl": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Required: true,
				Description: `Set of permissions associated with the key.
The possible ACLs are:
  - ` + "`search`" + `: allowed to perform search operations.
  - ` + "`browse`" + `: allowed to retrieve all index data with the browse endpoint.
  - ` + "`addObject`" + `: allowed to add or update a records in the index.
  - ` + "`deleteObject`" + `: allowed to delete an existing record.
  - ` + "`listIndexes`" + `: allowed to get a list of all existing indices.
  - ` + "`deleteIndex`" + `: allowed to delete an index.
  - ` + "`settings`" + `: allowed to read all index settings.
  - ` + "`editSettings`" + `: allowed to update all index settings.
  - ` + "`analytics`" + `: allowed to retrieve data with the Analytics API.
  - ` + "`recommendation`" + `: allowed to interact with the Recommendation API.
  - ` + "`usage`" + ` allowed to retrieve data with the Usage API.
  - ` + "`nluReadAnswers`" + `: allowed to perform semantic search with the Answers API.
  - ` + "`logs`" + `: allowed to query the logs.
  - ` + "`seeUnretrievableAttributes`" + `: allowed to retrieve unretrievableAttributes for all operations that return records.
`,
			},
			"expires_at": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Unix timestamp of the date at which the key expires. RFC3339 format. Will not expire per default.",
			},
			"max_hits_per_query": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Maximum number of hits this API key can retrieve in one call. This parameter can be used to protect you from attempts at retrieving your entire index contents by massively querying the index.",
			},
			"max_queries_per_ip_per_hour": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				Description: `Maximum number of API calls allowed from an IP address per hour.Each time an API call is performed with this key, a check is performed. If the IP at the source of the call did more than this number of calls in the last hour, a 429 code is returned.

This parameter can be used to protect you from attempts at retrieving your entire index contents by massively querying the index.`,
			},
			"indexes": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of targeted indices. You can target all indices starting with a prefix or ending with a suffix using the \u2018*\u2019 character. For example, \u201cdev_*\u201d matches all indices starting with \u201cdev_\u201d and \u201c*_dev\u201d matches all indices ending with \u201c_dev\u201d.",
			},
			"referers": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of referrers that can perform an operation. You can use the \"*\" (asterisk) character as a wildcard to match subdomains, or all pages of a website. For example, `\"https://algolia.com/\\*\"` matches all referrers starting with `\"https://algolia.com/\"`, and `\"\\*.algolia.com\"` matches all referrers ending with `\".algolia.com\"`. If you want to allow all possible referrers from the `algolia.com` domain, you can use `\"\\*algolia.com/\\*\"`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the API key.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The unix time at which the key has been created.",
			},
		},
	}
}

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	apiKey := mapToAPIKey(d)
	res, err := apiClient.searchClient.AddApiKey(
		apiClient.searchClient.NewApiAddApiKeyRequest(apiKey),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := waitForAPIKeyReady(ctx, apiClient, res.Key); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("key", res.Key); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

// waitForAPIKeyReady polls GetApiKey until the new key is readable, working around v4's
// WaitForApiKey(ADD) which treats any non-404 APIError (including 403 propagation
// responses) as "done" and exits prematurely.
func waitForAPIKeyReady(ctx context.Context, apiClient *apiClient, key string) error {
	return algoliautil.Poll(ctx, fmt.Sprintf("api key %q readiness", key), 60, func() (bool, error) {
		_, err := apiClient.searchClient.GetApiKey(
			apiClient.searchClient.NewApiGetApiKeyRequest(key),
			search.WithContext(ctx),
		)
		if err == nil {
			return true, nil
		}
		if algoliautil.IsNotFoundError(err) || algoliautil.IsForbiddenError(err) {
			return false, nil
		}
		return false, err
	})
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := refreshAPIKeyState(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	keyValue := d.Get("key").(string)
	apiKey := mapToAPIKey(d)
	_, err := apiClient.searchClient.UpdateApiKey(
		apiClient.searchClient.NewApiUpdateApiKeyRequest(keyValue, apiKey),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := waitForAPIKeyUpdated(ctx, apiClient, keyValue, apiKey); err != nil {
		return diag.FromErr(err)
	}

	return resourceAPIKeyRead(ctx, d, m)
}

// waitForAPIKeyUpdated polls GetApiKey until the static fields of the key match the
// requested update. We don't use v4's WaitForApiKey(UPDATE) because it compares
// `Validity` (remaining seconds), which decreases every second and never matches
// the value sent in the update request.
func waitForAPIKeyUpdated(ctx context.Context, apiClient *apiClient, key string, expected *search.ApiKey) error {
	return algoliautil.Poll(ctx, fmt.Sprintf("api key %q update", key), 60, func() (bool, error) {
		got, err := apiClient.searchClient.GetApiKey(
			apiClient.searchClient.NewApiGetApiKeyRequest(key),
			search.WithContext(ctx),
		)
		if err != nil {
			if algoliautil.IsNotFoundError(err) || algoliautil.IsForbiddenError(err) {
				return false, nil
			}
			return false, err
		}
		return apiKeyMatches(expected, got), nil
	})
}

func apiKeyMatches(expected *search.ApiKey, got *search.GetApiKeyResponse) bool {
	if expected.GetDescription() != got.GetDescription() {
		return false
	}
	if expected.GetMaxHitsPerQuery() != got.GetMaxHitsPerQuery() {
		return false
	}
	if expected.GetMaxQueriesPerIPPerHour() != got.GetMaxQueriesPerIPPerHour() {
		return false
	}
	if !stringSlicesEqualUnordered(aclsToStrings(expected.GetAcl()), aclsToStrings(got.GetAcl())) {
		return false
	}
	if !stringSlicesEqualUnordered(expected.GetIndexes(), got.GetIndexes()) {
		return false
	}
	if !stringSlicesEqualUnordered(expected.GetReferers(), got.GetReferers()) {
		return false
	}
	return true
}

func aclsToStrings(acls []search.Acl) []string {
	out := make([]string, len(acls))
	for i, a := range acls {
		out[i] = string(a)
	}
	return out
}

func stringSlicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := map[string]int{}
	for _, v := range a {
		counts[v]++
	}
	for _, v := range b {
		if counts[v] == 0 {
			return false
		}
		counts[v]--
	}
	return true
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiClient := m.(*apiClient)

	keyStr := d.Get("key").(string)
	_, err := apiClient.searchClient.DeleteApiKey(
		apiClient.searchClient.NewApiDeleteApiKeyRequest(keyStr),
		search.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err = apiClient.searchClient.WaitForApiKey(keyStr, search.API_KEY_OPERATION_DELETE, search.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAPIKeyStateContext(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("key", d.Id()); err != nil {
		return nil, err
	}

	if err := refreshAPIKeyState(ctx, d, m); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func refreshAPIKeyState(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*apiClient)

	keyID := d.Get("key").(string)
	key, err := apiClient.searchClient.GetApiKey(
		apiClient.searchClient.NewApiGetApiKeyRequest(keyID),
		search.WithContext(ctx),
	)
	if err != nil {
		if algoliautil.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("api key (%s) not found, removing from state", d.Id()))
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(strconv.FormatInt(key.CreatedAt, 10))

	acl := make([]string, len(key.Acl))
	for i, a := range key.Acl {
		acl[i] = string(a)
	}

	values := map[string]interface{}{
		"key":                         keyID,
		"acl":                         acl,
		"max_hits_per_query":          int(key.GetMaxHitsPerQuery()),
		"max_queries_per_ip_per_hour": int(key.GetMaxQueriesPerIPPerHour()),
		"referers":                    key.Referers,
		"description":                 key.GetDescription(),
		"indexes":                     key.Indexes,
		"created_at":                  key.CreatedAt,
	}
	// we can't set from key.Validity since it is remaining valid time and the value changes every second.
	// TODO: fix to work with import
	if expiresAtRFC3339, ok := d.GetOk("expires_at"); ok {
		values["expires_at"] = expiresAtRFC3339
	}
	if err := setValues(d, values); err != nil {
		return err
	}

	return nil
}

func mapToAPIKey(d *schema.ResourceData) *search.ApiKey {
	var validity *int32
	if expiresAtRFC3339, ok := d.GetOk("expires_at"); ok && expiresAtRFC3339 != "" {
		t, _ := time.Parse(time.RFC3339, expiresAtRFC3339.(string))
		v := int32(t.Unix() - time.Now().Unix())
		validity = &v
	}

	aclStrings := castStringSet(d.Get("acl"))
	acl := make([]search.Acl, len(aclStrings))
	for i, s := range aclStrings {
		acl[i] = search.Acl(s)
	}

	maxHitsPerQuery := int32(d.Get("max_hits_per_query").(int))
	maxQueriesPerIPPerHour := int32(d.Get("max_queries_per_ip_per_hour").(int))
	description := d.Get("description").(string)

	apiKey := &search.ApiKey{
		Acl:                    acl,
		Description:            &description,
		Indexes:                castStringSet(d.Get("indexes")),
		MaxHitsPerQuery:        &maxHitsPerQuery,
		MaxQueriesPerIPPerHour: &maxQueriesPerIPPerHour,
		Referers:               castStringSet(d.Get("referers")),
		Validity:               validity,
	}

	return apiKey
}
