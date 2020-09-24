package cache

import (
	"regexp"
)

// RegexCache is the default loading cache used for caching the compiled
// regex expression. If the default size is too big/small and/or a cache limit isn't
// desired at all, caller can simply replace the cache during global initialization.
// But be aware it's global so any packages uses this package inside your process will
// be affected.
var RegexCache = NewLoadingCache()

// GetRegex compiles a given regex pattern and returns a compiled *regexp.Regexp
// or error.
func GetRegex(pattern string) (*regexp.Regexp, error) {
	exp, err := RegexCache.Get(pattern, func(key interface{}) (interface{}, error) {
		return regexp.Compile(key.(string))
	})
	if err != nil {
		return nil, err
	}
	return exp.(*regexp.Regexp), nil
}
