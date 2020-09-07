package customfuncs

import (
	"fmt"
	"strconv"

	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

func avg(_ *transformctx.Ctx, values ...string) (string, error) {
	if len(values) == 0 {
		return "0", nil
	}
	s := float64(0)
	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", err
		}
		s += f
	}
	return fmt.Sprintf("%v", s/float64(len(values))), nil
}

func sum(_ *transformctx.Ctx, values ...string) (string, error) {
	s := float64(0)
	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "", err
		}
		s += f
	}
	return fmt.Sprintf("%v", s), nil
}
