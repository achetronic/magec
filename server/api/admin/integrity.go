//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package admin

import (
	"net/http"

	"github.com/achetronic/magec/server/store"
)

// ReferencesResponse is the 409 body returned when a delete target is still
// referenced by other entities. It is a demolition quote, not a refusal:
// repeating the DELETE with ?force=true scrubs every listed reference and
// proceeds.
type ReferencesResponse struct {
	Error      string            `json:"error" example:"entity is referenced by other entities"`
	References []store.Reference `json:"references"`
}

// DeadReferencesResponse lists references whose target no longer exists.
type DeadReferencesResponse struct {
	References []store.DeadReference `json:"references"`
}

// CleanDeadReferencesResponse reports how many dead references were removed.
type CleanDeadReferencesResponse struct {
	Removed int `json:"removed"`
}

// deleteGuard enforces referential integrity on entity deletes. When the
// entity is referenced and force is not requested it writes the 409 breakdown
// and returns false; with ?force=true it scrubs every reference first. A true
// return means the caller may proceed with the actual delete.
func (h *Handler) deleteGuard(w http.ResponseWriter, r *http.Request, id string) bool {
	refs := h.store.Referrers(id)
	if len(refs) == 0 {
		return true
	}
	if r.URL.Query().Get("force") != "true" {
		writeJSON(w, http.StatusConflict, ReferencesResponse{
			Error:      "entity is referenced by other entities",
			References: refs,
		})
		return false
	}
	if _, err := h.store.ScrubReferences(id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to scrub references: "+err.Error())
		return false
	}
	return true
}

// listDeadReferences returns every reference whose target no longer exists.
// @Summary      List dead references
// @Description  Returns references pointing at entities that no longer exist (left behind by deletes performed before referential integrity existed)
// @Tags         integrity
// @Produce      json
// @Success      200  {object}  DeadReferencesResponse
// @Security     AdminAuth
// @Router       /integrity/dead-references [get]
func (h *Handler) listDeadReferences(w http.ResponseWriter, r *http.Request) {
	dead := h.store.DeadReferences()
	if dead == nil {
		dead = []store.DeadReference{}
	}
	writeJSON(w, http.StatusOK, DeadReferencesResponse{References: dead})
}

// cleanDeadReferences scrubs every dead reference from the store.
// @Summary      Clean dead references
// @Description  Removes every reference pointing at entities that no longer exist
// @Tags         integrity
// @Produce      json
// @Success      200  {object}  CleanDeadReferencesResponse
// @Failure      500  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /integrity/dead-references/clean [post]
func (h *Handler) cleanDeadReferences(w http.ResponseWriter, r *http.Request) {
	removed, err := h.store.CleanDeadReferences()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, CleanDeadReferencesResponse{Removed: removed})
}
