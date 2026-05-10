package chain

import (
	"errors"
	"net/http"

	httperr "promptvault/internal/delivery/http/errors"
	chainuc "promptvault/internal/usecases/chain"
	quotauc "promptvault/internal/usecases/quota"
)

func respondError(w http.ResponseWriter, r *http.Request, err error) {
	var qe *quotauc.QuotaExceededError
	if errors.As(err, &qe) {
		httperr.RespondQuotaError(w, qe.QuotaType, qe.Used, qe.Limit, qe.PlanID, qe.Message)
		return
	}

	switch {
	case errors.Is(err, chainuc.ErrNotFound),
		errors.Is(err, chainuc.ErrStepNotFound),
		errors.Is(err, chainuc.ErrExecutionNotFound),
		errors.Is(err, chainuc.ErrPromptNotFound):
		httperr.Respond(w, httperr.NotFound(err.Error()))
	case errors.Is(err, chainuc.ErrForbidden),
		errors.Is(err, chainuc.ErrViewerReadOnly):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, chainuc.ErrInvalidName),
		errors.Is(err, chainuc.ErrInvalidDescription),
		errors.Is(err, chainuc.ErrInvalidVariableMapping),
		errors.Is(err, chainuc.ErrEmptyChain),
		errors.Is(err, chainuc.ErrInvalidStepPosition),
		errors.Is(err, chainuc.ErrReorderMismatch),
		errors.Is(err, chainuc.ErrInvalidConditions),
		errors.Is(err, chainuc.ErrInvalidNextStep),
		errors.Is(err, chainuc.ErrCycleInBranches),
		errors.Is(err, chainuc.ErrInvalidForkStep),
		errors.Is(err, chainuc.ErrInvalidBranchLabel),
		errors.Is(err, chainuc.ErrDuplicateBranchLabel),
		errors.Is(err, chainuc.ErrChooseBranchRequired),
		errors.Is(err, chainuc.ErrChosenBranchNotFound),
		errors.Is(err, chainuc.ErrCannotInsertAfterFork),
		errors.Is(err, chainuc.ErrParentNotFork),
		errors.Is(err, chainuc.ErrInsertForkLosesTail),
		errors.Is(err, chainuc.ErrCannotMoveFork),
		errors.Is(err, chainuc.ErrCannotMoveAtBoundary):
		httperr.Respond(w, httperr.BadRequest(err.Error()))
	case errors.Is(err, chainuc.ErrForkRequiresMax):
		httperr.Respond(w, httperr.Forbidden(err.Error()))
	case errors.Is(err, chainuc.ErrChainHasActiveExecutions),
		errors.Is(err, chainuc.ErrConcurrentAdvance):
		httperr.Respond(w, httperr.Conflict(err.Error()))
	case errors.Is(err, chainuc.ErrExecutionAlreadyCompleted):
		httperr.Respond(w, httperr.New(http.StatusUnprocessableEntity, err.Error(), nil))
	default:
		httperr.RespondWithRequest(w, r, httperr.Internal(err))
	}
}
