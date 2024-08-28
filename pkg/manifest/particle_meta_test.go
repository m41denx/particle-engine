package manifest

import (
	"golang.org/x/exp/slices"
	"testing"
)

var TestCases = map[string][]string{
	"https://domain.com/user/particle@tag-22": {"https://domain.com/", "user", "particle", "tag-22"},
	"http://domain.com/user/particle@tag-22":  {"http://domain.com/", "user", "particle", "tag-22"},
	"domain.com/user/particle@tag-22":         nil,
	"user/particle@tag-22":                    {"http://particles.fruitspace.one/", "user", "particle", "tag-22"},
	"user/particle":                           {"http://particles.fruitspace.one/", "user", "particle", "latest"},
	"user/particle@.2.2.":                     {"http://particles.fruitspace.one/", "user", "particle", ".2.2."},
	"user/particle@.2.2.@2.2.2":               nil,
}

func TestParseParticleURL(t *testing.T) {
	for tcase, tres := range TestCases {
		val, err := ParseParticleURL(tcase)
		if err != nil {
			if tres == nil {
				continue
			}
			t.Errorf("%s: %v", tcase, err)
			continue
		}
		expected := []string{val.Server, val.User, val.Name, val.Tag}
		if !slices.Equal(tres, expected) {
			t.Errorf("expected %v, got %v", tres, expected)
		}
	}
}
