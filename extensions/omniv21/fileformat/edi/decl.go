package edi

type fileDecl struct {
	SegDelim    string     `json:"segment_delimiter,omitempty"`
	ElemDelim   string     `json:"element_delimiter,omitempty"`
	CompDelim   *string    `json:"component_delimiter,omitempty"`
	ReleaseChar *string    `json:"release_character,omitempty"`
	IgnoreCRLF  bool       `json:"ignore_crlf,omitempty"`
	SegDecls    []*segDecl `json:"segment_declarations,omitempty"`
}
