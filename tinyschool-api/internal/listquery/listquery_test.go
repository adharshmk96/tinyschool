package listquery

import (
	"net/url"
	"testing"
)

type item struct {
	name string
}

func TestParseDefaults(t *testing.T) {
	options, err := Parse(url.Values{}, map[string]bool{"name": true}, "name")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if options.Page != 1 || options.PageSize != DefaultPageSize || options.Order != "asc" || options.Sort != "name" {
		t.Fatalf("unexpected defaults: %#v", options)
	}
}

func TestParseRejectsInvalidValues(t *testing.T) {
	tests := []url.Values{
		{"page": {"0"}},
		{"page": {"abc"}},
		{"pageSize": {"0"}},
		{"pageSize": {"101"}},
		{"order": {"sideways"}},
		{"sort": {"unknown"}},
	}
	for _, values := range tests {
		if _, err := Parse(values, map[string]bool{"name": true}, "name"); err == nil {
			t.Errorf("Parse(%v) expected error", values)
		}
	}
}

func TestApplyFiltersSortsAndPaginates(t *testing.T) {
	items := []item{{name: "Charlie"}, {name: "Alpine"}, {name: "Alpha"}}
	result := Apply(items, Options{Search: "al", Sort: "name", Order: "desc", Page: 2, PageSize: 1},
		func(value item, search string) bool { return Contains(search, value.name) },
		func(a, b item, _ string) bool { return a.name < b.name },
	)
	if result.Total != 2 || len(result.Items) != 1 || result.Items[0].name != "Alpha" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
