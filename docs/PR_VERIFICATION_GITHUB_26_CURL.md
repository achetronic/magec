# Manual verification: GitHub #26 (`snake_case` on `POST /agent/run`)

Reproducible **curl + jq** flow with **environment variables** (no hard-coded UUIDs).  
（繁中說明：變數占位、可重現的驗收步驟，供 PR 描述或本地對照。）

下列步驟使用**變數占位**，可在 bash／Git Bash／WSL 一鍵複製調整。需已啟動 Magec（預設 user API `:8080`、admin API `:8081`），並有一個可連線的 **OpenAI 相容**後端（例如 Ollama：`http://localhost:11434/v1`；若 Magec 在 Docker 內，後端 URL 常用 `http://host.docker.internal:11434/v1`）。

**前置：**安裝 `curl` 與 `jq`。

```bash
# --- 依你的環境修改 ---
export MAGE_USER_BASE="http://127.0.0.1:8080"   # User API（/api/v1/agent/...）
export MAGE_ADMIN_BASE="http://127.0.0.1:8081"  # Admin API（/api/v1/admin/...）
export ADMIN_PASSWORD="your-admin-password"     # 對應 config server.adminPassword

# 後端：OpenAI 相容 Base URL（勿省略 /v1）
export LLM_BASE_URL="http://host.docker.internal:11434/v1"

# Ollama 上實際存在的模型名
export LLM_MODEL="qwen3:8b"

# 之後由 API 填入
export CLIENT_TOKEN=""
export BACKEND_ID=""
export AGENT_ID=""
export CLIENT_ID=""
```

## 1) 取得 Direct client 的 token 與 id

若你已有一個 `type=direct` 的 client，從列表抄 `id` 與 `token` 即可：

```bash
curl -sS "${MAGE_ADMIN_BASE}/api/v1/admin/clients" \
  -H "Authorization: Bearer ${ADMIN_PASSWORD}" | jq .
```

手動設定（或由 jq 篩選第一個 direct）：

```bash
CLIENT_ID="$(curl -sS "${MAGE_ADMIN_BASE}/api/v1/admin/clients" \
  -H "Authorization: Bearer ${ADMIN_PASSWORD}" | jq -r '.[] | select(.type=="direct") | .id' | head -1)"

CLIENT_TOKEN="$(curl -sS "${MAGE_ADMIN_BASE}/api/v1/admin/clients" \
  -H "Authorization: Bearer ${ADMIN_PASSWORD}" | jq -r --arg id "$CLIENT_ID" '.[] | select(.id==$id) | .token')"
```

## 2) 建立 Backend（OpenAI 相容）

```bash
BACKEND_ID="$(
  curl -sS -X POST "${MAGE_ADMIN_BASE}/api/v1/admin/backends" \
    -H "Authorization: Bearer ${ADMIN_PASSWORD}" \
    -H "Content-Type: application/json" \
    -d "$(jq -n \
      --arg url "$LLM_BASE_URL" \
      '{name:"verify-gh26-backend",type:"openai",url:$url}')" \
  | jq -r '.id'
)"
echo "BACKEND_ID=${BACKEND_ID}"
```

## 3) 建立 Agent（綁定上述 backend + 模型）

```bash
AGENT_ID="$(
  curl -sS -X POST "${MAGE_ADMIN_BASE}/api/v1/admin/agents" \
    -H "Authorization: Bearer ${ADMIN_PASSWORD}" \
    -H "Content-Type: application/json" \
    -d "$(jq -n \
      --arg b "$BACKEND_ID" \
      --arg m "$LLM_MODEL" \
      '{name:"verify-gh26-agent",llm:{backend:$b,model:$m}}')" \
  | jq -r '.id'
)"
echo "AGENT_ID=${AGENT_ID}"
```

## 4) 讓該 client 能使用此 agent（`allowedAgents`）

```bash
curl -sS -X PUT "${MAGE_ADMIN_BASE}/api/v1/admin/clients/${CLIENT_ID}" \
  -H "Authorization: Bearer ${ADMIN_PASSWORD}" \
  -H "Content-Type: application/json" \
  -d "$(curl -sS "${MAGE_ADMIN_BASE}/api/v1/admin/clients/${CLIENT_ID}" \
        -H "Authorization: Bearer ${ADMIN_PASSWORD}" \
      | jq --arg a "$AGENT_ID" '.allowedAgents = [$a]')" | jq .
```

（若你希望保留原有 allowed 清單，請自行改 jq 合併陣列。）

## 5) 確認 `list-apps` 可見該 agent

```bash
curl -sS "${MAGE_USER_BASE}/api/v1/agent/list-apps" \
  -H "Authorization: Bearer ${CLIENT_TOKEN}" | jq .
# 預期：JSON 陣列中含 "${AGENT_ID}"
```

## 6) 建立 session（ADK 慣例：先建 session 再 run）

自訂 user／session 字串（僅需全程一致）：

```bash
export GH26_USER="gh26-verify-user"
export GH26_SESSION="gh26-verify-session"

curl -sS -X POST \
  "${MAGE_USER_BASE}/api/v1/agent/apps/${AGENT_ID}/users/${GH26_USER}/sessions/${GH26_SESSION}" \
  -H "Authorization: Bearer ${CLIENT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
# 預期：HTTP 200
```

## 7) 驗收重點：`POST /run` 僅送 **snake_case**（#26）

```bash
curl -sS -w "\nHTTP %{http_code}\n" --max-time 120 -X POST \
  "${MAGE_USER_BASE}/api/v1/agent/run" \
  -H "Authorization: Bearer ${CLIENT_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "$(jq -n \
    --arg app "$AGENT_ID" \
    --arg u "$GH26_USER" \
    --arg s "$GH26_SESSION" \
    '{app_name:$app,user_id:$u,session_id:$s,new_message:{role:"user",parts:[{text:"Reply with exactly one word: pong"}]}}')"
```

**預期：**HTTP **200**，回應 JSON 內模型文字含 **`pong`**（或語意等同之回覆）。

若此時仍出現 **`session … not found`**，且已確認步驟 6 為 200、步驟 7 的 `app_name`／`user_id`／`session_id` 與步驟 6 路徑一致，則再對照本 PR 的 `RunAgentJSONNormalize` 是否已套在 `/api/v1/agent/` 鏈上。

## 8) 自動化測試（無需真 LLM）

```bash
cd server
go test ./middleware/ -run 'RunAgentJSONNormalize|Normalize' -count=1 -v
```

---

## 選用：同流程驗 `run_sse`

將步驟 7 的 URL 改為 `"${MAGE_USER_BASE}/api/v1/agent/run_sse"`，並依客戶端能力處理 `text/event-stream`（本文件略）。
