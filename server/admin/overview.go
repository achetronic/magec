package admin

import (
	"net/http"
)

// getOverview returns a dashboard summary of all resources.
// @Summary      Dashboard overview
// @Description  Returns counts for all resource types and agent summaries
// @Tags         overview
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /overview [get]
func (h *Handler) getOverview(w http.ResponseWriter, r *http.Request) {
	data := h.store.Data()

	agentSummaries := make([]map[string]interface{}, len(data.Agents))
	for i, a := range data.Agents {
		agentSummaries[i] = map[string]interface{}{
			"id":          a.ID,
			"name":        a.Name,
			"description": a.Description,
			"llmBackend":  a.LLM.Backend,
			"llmModel":    a.LLM.Model,
			"mcpCount":    len(a.MCPServers),
			"telegram":    a.Telegram.Enabled,
		}
	}

	overview := map[string]interface{}{
		"backendsCount":        len(data.Backends),
		"memoryProvidersCount": len(data.MemoryProviders),
		"mcpServersCount":      len(data.MCPServers),
		"agentsCount":          len(data.Agents),
		"devicesCount":         len(data.Devices),
		"cronJobsCount":        len(data.CronJobs),
		"agents":               agentSummaries,
	}

	writeJSON(w, http.StatusOK, overview)
}
