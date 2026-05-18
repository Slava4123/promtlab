package prompt_insights

import "errors"

var (
	// ErrUnknownInsightType — handler передал тип, которого нет в whitelist.
	ErrUnknownInsightType = errors.New("unknown insight type")

	// ErrPromptsNotOwned — юзер запрашивает merge промптов, которые ему не принадлежат.
	ErrPromptsNotOwned = errors.New("prompts not owned by user")

	// ErrSamePrompt — merge id1 == id2.
	ErrSamePrompt = errors.New("cannot merge prompt with itself")

	// ErrProRequired — план юзера не разрешает данный insight type.
	// Маппится в HTTP 402 (как в usecases/analytics).
	ErrProRequired = errors.New("pro plan required for this insight")
)
