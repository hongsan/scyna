package scyna_data

import (
	"log"
	"scyna"

	"github.com/scylladb/gocqlx/v2/qb"
)

type ModuleType int32

const (
	PUBLIC  ModuleType = 1
	PRIVATE ModuleType = 2
)

type Organization struct {
	Code string
}

func NewOrganization(code string, name string, password string) *Organization {
	if applied, err := qb.Insert("scyna.organization").
		Columns("code", "name", "password").
		Unique().Query(scyna.DB).
		Bind(code, name, password).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in create Organization:", err)
		}
	}
	return &Organization{Code: code}
}

func (org *Organization) AddModule(code string, description string, t ModuleType) *Organization {
	if applied, err := qb.Insert("scyna.module").
		Columns("code", "description", "type", "org_code").
		Unique().Query(scyna.DB).
		Bind(code, description, uint16(t), org.Code).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in AddModule:", err)
		}
	}

	if applied, err := qb.Insert("scyna.org_has_module").
		Columns("org_code", "module_code").
		Unique().Query(scyna.DB).
		Bind(org.Code, code).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in create org_has_module:", err)
		}
	}

	return org
}

func (org *Organization) AddApplication(code string, name string, auth string) *Organization {
	if applied, err := qb.Insert("scyna.application").
		Columns("code", "name", "auth", "org_code").
		Unique().Query(scyna.DB).
		Bind(code, name, auth, org.Code).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in AddApplication")
		}
	}

	if applied, err := qb.Insert("scyna.org_has_app").
		Columns("org_code", "app_code").
		Unique().Query(scyna.DB).
		Bind(org.Code, code).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in create org_has_app")
		}
	}
	return org
}

func (org *Organization) AddClient(id string, secret string) *Organization {
	if applied, err := qb.Insert("scyna.client").
		Columns("id", "secret", "org_code").
		Unique().Query(scyna.DB).
		Bind(id, secret, org.Code).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in AddClient")
		}
	}

	if applied, err := qb.Insert("scyna.org_has_client").
		Columns("org_code", "client_id").
		Unique().Query(scyna.DB).
		Bind(org.Code, id).ExecCASRelease(); !applied {
		if err != nil {
			log.Fatal("Error in create org_has_client")
		}
	}
	return org
}
