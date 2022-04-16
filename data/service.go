package scyna_data

import (
	"log"
	"scyna"

	"github.com/scylladb/gocqlx/v2/qb"
)

type Service struct {
	URL string
}

func CreateService(module string, url string) *Service {
	if applied, err := qb.Insert("scyna.service").
		Columns("module_code", "url").
		Unique().Query(scyna.DB).
		Bind(module, url).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in create Service")
		}
	}
	return &Service{URL: url}
}

func (s *Service) AttachToApp(app string) *Service {
	if applied, err := qb.Insert("scyna.app_use_service").
		Columns("app_code", "service_url").
		Unique().Query(scyna.DB).
		Bind(app, s.URL).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in AttatchToApp")
		}
	}
	return s
}

func (s *Service) AttachToClient(clientid string) *Service {
	if applied, err := qb.Insert("scyna.client_use_service").
		Columns("client_id", "service_url").
		Unique().Query(scyna.DB).
		Bind(clientid, s.URL).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in AttatchToClient")
		}
	}
	return s
}
