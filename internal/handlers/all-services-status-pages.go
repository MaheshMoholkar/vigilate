package handlers

import (
	"net/http"

	"github.com/tsawler/vigilate/internal/helpers"

	"github.com/CloudyKit/jet/v6"
)

// AllHealthyServices lists all healthy services
func (repo *DBRepo) AllHealthyServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("healthy")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "healthy", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllWarningServices lists all warning services
func (repo *DBRepo) AllWarningServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("warning")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "warning", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllProblemServices lists all problem services
func (repo *DBRepo) AllProblemServices(w http.ResponseWriter, r *http.Request) {
	services, err := repo.DB.GetServicesByStatus("problem")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "problems", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}

// AllPendingServices lists all pending services
func (repo *DBRepo) AllPendingServices(w http.ResponseWriter, r *http.Request) {
	// get all host services (with host info) for status pending
	services, err := repo.DB.GetServicesByStatus("pending")
	if err != nil {
		printTemplateError(w, err)
		return
	}

	vars := make(jet.VarMap)
	vars.Set("services", services)

	err = helpers.RenderPage(w, r, "pending", vars, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
