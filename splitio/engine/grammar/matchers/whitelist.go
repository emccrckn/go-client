package matchers

import (
	"github.com/splitio/go-toolkit/datastructures/set"
)

// WhitelistMatcher matches if the key received is present in the matcher's whitelist
type WhitelistMatcher struct {
	Matcher
	whitelist *set.ThreadUnsafeSet
}

// Match returns true if the key is present in the whitelist.
func (m *WhitelistMatcher) Match(key string, attributes map[string]interface{}, bucketingKey *string) bool {
	matchingKey, err := m.matchingKey(key, attributes)
	stringMatchingKey, ok := matchingKey.(string)
	return err == nil && ok && m.whitelist.Has(stringMatchingKey)
}

// NewWhitelistMatcher returns a new WhitelistMatcher
func NewWhitelistMatcher(negate bool, whitelist []string, attributeName *string) *WhitelistMatcher {
	wlSet := set.NewSet()
	for _, elem := range whitelist {
		wlSet.Add(elem)
	}
	return &WhitelistMatcher{
		Matcher: Matcher{
			negate:        negate,
			attributeName: attributeName,
		},
		whitelist: wlSet,
	}
}
