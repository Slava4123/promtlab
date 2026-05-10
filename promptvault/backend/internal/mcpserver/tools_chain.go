package mcpserver

import (
	"context"
	"encoding/json"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	authmw "promptvault/internal/middleware/auth"
)

// Phase 16: Prompt Chains в MCP. Loop работает так:
//   1. start_chain_execution(chain_id, initial_vars) → ExecutionState (current_step, snapshot)
//   2. MCP-клиент сам вызывает LLM с current_step rendered prompt'ом
//   3. advance_chain_step(execution_id, step_output) → следующий ExecutionState или status='completed'
//   4. Повторять п.2-3 пока status != 'completed'.

type ListChainsInput struct {
	TeamID *uint `json:"team_id,omitempty" jsonschema:"Team ID (omit for personal workspace)"`
	Limit  int   `json:"limit,omitempty" jsonschema:"Max chains per page (default 20, max 100)"`
	Offset int   `json:"offset,omitempty" jsonschema:"Offset for pagination"`
}

type GetChainInput struct {
	ChainID uint `json:"chain_id" jsonschema:"Chain ID to retrieve"`
}

type StartChainExecutionInput struct {
	ChainID     uint            `json:"chain_id" jsonschema:"Chain ID to execute"`
	InitialVars json.RawMessage `json:"initial_vars,omitempty" jsonschema:"Optional initial variables (JSON object) for the chain run"`
}

type AdvanceChainStepInput struct {
	ExecutionID       uint   `json:"execution_id" jsonschema:"Execution ID returned by start_chain_execution"`
	StepOutput        string `json:"step_output" jsonschema:"Output text from LLM for the current step"`
	ChosenBranchIndex *int   `json:"chosen_branch_index,omitempty" jsonschema:"For fork steps: 0-based index of the branch to follow. Omit for prompt steps."`
}

func (t *toolHandlers) listChains(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListChainsInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "list_chains", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := enforceTeamID(ctx, input.TeamID); err != nil {
		return nil, nil, mapDomainError(err)
	}
	userID := authmw.GetUserID(ctx)

	var teamIDs []uint
	if input.TeamID != nil {
		teamIDs = []uint{*input.TeamID}
	}
	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	chains, total, err := t.chains.List(ctx, userID, teamIDs, limit, input.Offset)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}

	type chainSummary struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		TeamID      *uint  `json:"team_id,omitempty"`
	}
	items := make([]chainSummary, len(chains))
	for i, c := range chains {
		items[i] = chainSummary{ID: c.ID, Name: c.Name, Description: c.Description, TeamID: c.TeamID}
	}
	res, err := jsonResult(map[string]any{"items": items, "total": total, "limit": limit, "offset": input.Offset})
	return res, nil, err
}

func (t *toolHandlers) getChain(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetChainInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "get_chain", false); err != nil {
		return nil, nil, mapDomainError(err)
	}
	userID := authmw.GetUserID(ctx)
	c, err := t.chains.GetByIDWithSteps(ctx, input.ChainID, userID)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	res, err := jsonResult(c)
	return res, nil, err
}

func (t *toolHandlers) startChainExecution(ctx context.Context, _ *sdkmcp.CallToolRequest, input StartChainExecutionInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "start_chain_execution", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	userID := authmw.GetUserID(ctx)
	exec, err := t.chains.StartExecution(ctx, input.ChainID, userID, input.InitialVars)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.incrementMCPUsage(ctx)
	res, err := jsonResult(exec)
	return res, nil, err
}

func (t *toolHandlers) advanceChainStep(ctx context.Context, _ *sdkmcp.CallToolRequest, input AdvanceChainStepInput) (*sdkmcp.CallToolResult, any, error) {
	if err := enforceScope(ctx, "advance_chain_step", true); err != nil {
		return nil, nil, mapDomainError(err)
	}
	if err := t.checkMCPQuota(ctx); err != nil {
		return nil, nil, mapDomainError(err)
	}
	userID := authmw.GetUserID(ctx)
	exec, err := t.chains.AdvanceStep(ctx, input.ExecutionID, userID, input.StepOutput, input.ChosenBranchIndex)
	if err != nil {
		return nil, nil, mapDomainError(err)
	}
	t.incrementMCPUsage(ctx)
	res, err := jsonResult(exec)
	return res, nil, err
}

// registerChainTools — вызывается из NewMCPServer только если chains != nil.
// Опциональность позволяет отключать фичу через конфиг (Phase 16: feature flag).
func (t *toolHandlers) registerChainTools(server *sdkmcp.Server) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "list_chains",
		Title:       "List prompt chains",
		Description: "List prompt chains (linear sequences of prompts where each step's output becomes input for the next). Use chain.id to start_chain_execution.",
		Annotations: readOnlyAnnotations,
	}, t.listChains)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "get_chain",
		Title:       "Get chain detail",
		Description: "Get a chain with all its steps and variable mappings. Use to preview structure before start_chain_execution.",
		Annotations: readOnlyAnnotations,
	}, t.getChain)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "start_chain_execution",
		Title:       "Start chain execution",
		Description: "Start a new run of a prompt chain. Returns ExecutionState with current_step (1-based) and chain_snapshot (frozen structure + prompt contents). After receiving LLM output for current step, call advance_chain_step.",
		Annotations: writeAnnotations,
	}, t.startChainExecution)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "advance_chain_step",
		Title:       "Advance chain step",
		Description: "Submit LLM output for the current step and advance to the next. Returns updated ExecutionState. When status='completed', the chain is finished and step_outputs contains all results keyed by step_<id>.",
		Annotations: writeAnnotations,
	}, t.advanceChainStep)
}
