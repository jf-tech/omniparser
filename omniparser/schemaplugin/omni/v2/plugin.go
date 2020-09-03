package omniv2

import (
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin/omni/v2/transform"
)

const (
	pluginVersion = "omni.2.0"
	fileFormatXML = "xml"
)

func ParseSchema(_ *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
	return nil, errs.ErrSchemaNotSupported
}

type omniSchema struct {
	schemaplugin.Header
	Decls map[string]*transform.Decl `json:"transform_declarations"`
}
