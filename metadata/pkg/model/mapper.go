package model

import "github.com/ochamekan/ms/gen"

// MetadataToProto converts a Metadata struct into generated proto counterpart.
func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}

// MetadataFromProto converts a generate proto counterpart
// into a metadata struct.
func MetadataFromProto(m *gen.Metadata) *Metadata {
	return &Metadata{
		ID:          m.Id,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}
