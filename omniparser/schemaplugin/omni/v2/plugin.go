package omniv2

import (
	"github.com/jf-tech/omniparser/omniparser/errs"
	"github.com/jf-tech/omniparser/omniparser/schemaplugin"
)

const (
	pluginVersion = "omni.2.0"
	fileFormatXML = "xml"
)

// ParseSchema parses, validates and creates an omni-schema based schema plugin.
func ParseSchema(_ *schemaplugin.ParseSchemaCtx) (schemaplugin.Plugin, error) {
	return nil, errs.ErrSchemaNotSupported
}

type schemaPlugin struct{}
