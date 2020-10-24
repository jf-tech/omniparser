package edi

type fileDecl struct {
	SegDelim    string     `json:"segment_delimiter,omitempty"`
	ElemDelim   string     `json:"element_delimiter,omitempty"`
	CompDelim   *string    `json:"component_delimiter,omitempty"`
	ReleaseChar *string    `json:"release_character,omitempty"`
	SegDecls    []*segDecl `json:"segment_declarations,omitempty"`
}
