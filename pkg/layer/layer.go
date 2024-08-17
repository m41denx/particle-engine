package layer

const DefaultLayerRepo = "http://particles.fruitspace.one/layers/"

type Layer struct {
	Hash      string
	Files     []string
	Deletions []string

	//For download
	dir    string
	server string
}

func NewLayer(hash string, dir string, server string) *Layer {
	if server == "" {
		server = DefaultLayerRepo
	}
	return &Layer{
		Hash:      hash,
		Files:     []string{},
		Deletions: []string{},
		dir:       dir,
		server:    server,
	}
}
