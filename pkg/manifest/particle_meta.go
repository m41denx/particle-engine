package manifest

import (
	"fmt"
	"regexp"
)

type ParticleMeta struct {
	Server   string
	User     string
	Name     string
	Fullname string
	Tag      string
}

func ParseParticleURL(url string) (ParticleMeta, error) {
	meta := ParticleMeta{}
	if url == "blank" {
		return ParticleMeta{
			Server:   DefaultRepo,
			User:     "",
			Name:     "blank",
			Fullname: "blank",
			Tag:      "",
		}, nil
	}
	re := regexp.MustCompile("^(?P<server>http(s)?://[a-zA-Z0-9_./-]+/)?(?P<user>[a-z0-9-]+)/(?P<name>[a-z0-9-_.]+)(?P<tag>@[a-z0-9.-]+)?$")

	matches := re.FindStringSubmatch(url)

	if len(matches) == 0 {
		return meta, fmt.Errorf("invalid particle defenition: %s", url)
	}
	if len(matches[1]) > 0 {
		meta.Server = matches[1]
	} else {
		meta.Server = DefaultRepo
	}
	meta.User = matches[3]
	meta.Name = matches[4]
	if len(matches[5]) > 0 {
		meta.Tag = matches[5][1:]
	} else {
		meta.Tag = "latest"
	}
	meta.Fullname = fmt.Sprintf("%s/%s@%s", meta.User, meta.Name, meta.Tag)
	return meta, nil
}
