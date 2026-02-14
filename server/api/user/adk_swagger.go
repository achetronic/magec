package user

// ADK endpoint documentation stubs.
// These endpoints are served by the Google ADK router, not by our handlers.
// The annotations here exist solely to include them in the Swagger spec.

// RunAgent executes an agent synchronously.
// @Summary      Run agent
// @Description  Send a message to an agent and receive the full response. The app_name is the agent ID.
// @Tags         agent
// @Accept       json
// @Produce      json
// @Param        body  body      object  true  "Run request with app_name, user_id, session_id, and new_message"
// @Success      200   {object}  object  "Agent response"
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/run [post]
func (h *Handler) RunAgent() {}

// RunAgentSSE executes an agent with streaming via Server-Sent Events.
// @Summary      Run agent (SSE streaming)
// @Description  Send a message to an agent and receive the response streamed as Server-Sent Events.
// @Tags         agent
// @Accept       json
// @Produce      text/event-stream
// @Param        body  body      object  true  "Run request with app_name, user_id, session_id, and new_message"
// @Success      200   {object}  object  "SSE event stream"
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/run_sse [post]
func (h *Handler) RunAgentSSE() {}

// ListApps lists available agents.
// @Summary      List agents
// @Description  Returns a list of all available agent app names.
// @Tags         agent
// @Produce      json
// @Success      200  {array}   string  "List of app names"
// @Security     BearerAuth
// @Router       /agent/list-apps [get]
func (h *Handler) ListApps() {}

// ListSessions lists all sessions for a user-agent pair.
// @Summary      List sessions
// @Description  Returns all conversation sessions for the given app and user.
// @Tags         sessions
// @Produce      json
// @Param        app   path      string  true  "Agent app name (agent ID)"
// @Param        user  path      string  true  "User ID"
// @Success      200   {array}   object  "List of sessions"
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions [get]
func (h *Handler) ListSessions() {}

// CreateSession creates a new conversation session.
// @Summary      Create session
// @Description  Creates a new session for the given app and user. Returns the session object with its generated ID.
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        app   path      string  true  "Agent app name (agent ID)"
// @Param        user  path      string  true  "User ID"
// @Param        body  body      object  false "Optional session configuration"
// @Success      200   {object}  object  "Created session"
// @Failure      400   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions [post]
func (h *Handler) CreateSession() {}

// GetSession returns a session by ID.
// @Summary      Get session
// @Description  Returns the session details including message history.
// @Tags         sessions
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Success      200      {object}  object  "Session details"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session} [get]
func (h *Handler) GetSession() {}

// UpdateSession updates an existing session.
// @Summary      Update session
// @Description  Updates session data (e.g. appending events or modifying state).
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Param        body     body      object  true  "Session update payload"
// @Success      200      {object}  object  "Updated session"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session} [post]
func (h *Handler) UpdateSession() {}

// DeleteSession deletes a session.
// @Summary      Delete session
// @Description  Permanently deletes a conversation session and its history.
// @Tags         sessions
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Success      200      {string}  string  "Deleted"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session} [delete]
func (h *Handler) DeleteSession() {}

// ListArtifacts lists artifacts in a session.
// @Summary      List artifacts
// @Description  Returns all artifacts generated by agents during the given session.
// @Tags         artifacts
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Success      200      {array}   object  "List of artifacts"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session}/artifacts [get]
func (h *Handler) ListArtifacts() {}

// GetArtifact returns the latest version of an artifact.
// @Summary      Get artifact
// @Description  Returns the latest version of the named artifact from the session.
// @Tags         artifacts
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Param        name     path      string  true  "Artifact name"
// @Success      200      {object}  object  "Artifact data"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session}/artifacts/{name} [get]
func (h *Handler) GetArtifact() {}

// GetArtifactVersion returns a specific version of an artifact.
// @Summary      Get artifact version
// @Description  Returns a specific version of the named artifact.
// @Tags         artifacts
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Param        name     path      string  true  "Artifact name"
// @Param        version  path      string  true  "Version number"
// @Success      200      {object}  object  "Artifact data"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session}/artifacts/{name}/versions/{version} [get]
func (h *Handler) GetArtifactVersion() {}

// DeleteArtifact deletes an artifact.
// @Summary      Delete artifact
// @Description  Permanently deletes the named artifact and all its versions from the session.
// @Tags         artifacts
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Param        name     path      string  true  "Artifact name"
// @Success      200      {string}  string  "Deleted"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session}/artifacts/{name} [delete]
func (h *Handler) DeleteArtifact() {}

// GetTrace returns the trace for a given event.
// @Summary      Get event trace
// @Description  Returns the execution trace for a specific event, useful for debugging agent behavior.
// @Tags         debug
// @Produce      json
// @Param        event_id  path      string  true  "Event ID"
// @Success      200       {object}  object  "Trace data"
// @Failure      404       {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/debug/trace/{event_id} [get]
func (h *Handler) GetTrace() {}

// GetEventGraph returns the execution graph for an event within a session.
// @Summary      Get event graph
// @Description  Returns the execution graph showing agent interactions for a specific event.
// @Tags         debug
// @Produce      json
// @Param        app      path      string  true  "Agent app name (agent ID)"
// @Param        user     path      string  true  "User ID"
// @Param        session  path      string  true  "Session ID"
// @Param        event    path      string  true  "Event ID"
// @Success      200      {object}  object  "Event graph"
// @Failure      404      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /agent/apps/{app}/users/{user}/sessions/{session}/events/{event}/graph [get]
func (h *Handler) GetEventGraph() {}
