package optimizer

import "strings"

// Fields to always strip regardless of tool — ported from leanmcp/optimizers/generic.js
var alwaysStrip = map[string]bool{
	"self": true, "node_id": true, "gravatar_id": true,
	"avatar": true, "avatarUrls": true, "avatar_url": true,
	"iconUrl": true, "icon_url": true, "icon": true,
	"expand": true, "schema": true, "renderedFields": true,
	"versionedRepresentations": true, "editmeta": true,
	"operations": true, "names": true,
	"author_association": true, "active_lock_reason": true,
	"auto_merge": true, "merge_commit_sha": true,
	"performed_via_github_app": true,
	"timeline_url": true, "repository_url": true,
	"events_url": true, "commits_url": true,
	"review_comments_url": true, "review_comment_url": true,
	"comments_url": true, "statuses_url": true,
	"diff_url": true, "patch_url": true, "issue_url": true,
	"followers_url": true, "following_url": true,
	"gists_url": true, "starred_url": true,
	"subscriptions_url": true, "organizations_url": true,
	"repos_url": true, "received_events_url": true,
	"color": true, "initials": true,
	"profilePicture": true, "gravatar": true,
	"_links": true,
}

// URL fields to keep (human-facing links)
var keepURLFields = map[string]bool{
	"url": true, "html_url": true, "web_url": true,
	"webUrl": true, "permalink": true,
}

func isURLField(key string) bool {
	lower := strings.ToLower(key)
	if keepURLFields[key] {
		return false
	}
	return strings.HasSuffix(lower, "_url") ||
		strings.HasSuffix(lower, "url") && lower != "url" ||
		strings.HasSuffix(lower, "uri")
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	case float64:
		return false
	case bool:
		return false
	}
	return false
}

func stripNulls(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			if isEmpty(val) {
				continue
			}
			cleaned := stripNulls(val)
			if !isEmpty(cleaned) {
				result[key] = cleaned
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			if isEmpty(item) {
				continue
			}
			result = append(result, stripNulls(item))
		}
		return result
	default:
		return data
	}
}

func stripURLFields(data interface{}) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if alwaysStrip[k] {
			continue
		}
		if isURLField(k) {
			continue
		}
		switch child := v.(type) {
		case map[string]interface{}:
			result[k] = stripURLFields(child)
		case []interface{}:
			arr := make([]interface{}, 0, len(child))
			for _, item := range child {
				arr = append(arr, stripURLFields(item))
			}
			result[k] = arr
		default:
			result[k] = v
		}
	}
	return result
}

func stripFields(m map[string]interface{}, fields []string) map[string]interface{} {
	skip := make(map[string]bool, len(fields))
	for _, f := range fields {
		skip[f] = true
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if !skip[k] {
			result[k] = v
		}
	}
	return result
}
