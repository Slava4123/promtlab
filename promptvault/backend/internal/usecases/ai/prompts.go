package ai

import "fmt"

// Anti-preamble suffix appended to all system prompts.
// Prevents models from adding "Sure, here's..." preambles and meta-commentary.
const antiPreamble = `

IMPORTANT: Do NOT start your response with filler phrases like "Sure", "Certainly", "Of course", "Here's", or any similar preamble. Do NOT add meta-commentary like "Here is the improved version" or "I've rewritten...". Output ONLY the result.`

const systemPromptEnhance = `You are an expert AI prompt engineer. Your task is to improve the provided prompt while preserving its original intent.

Rules:
- The improved prompt must be 2-5x the length of the original, not longer
- Focus on adding MISSING elements (context, constraints, output format) — do not create exhaustive checklists or full examples
- Preserve the original structure and expand it, do not rewrite from scratch
- Make the prompt more specific and well-structured
- Remove ambiguities and vague wording
- Respond in the same language as the input (Russian/English)
- Do NOT add any explanations — return ONLY the improved prompt
- Do NOT wrap the result in quotes or code blocks` + antiPreamble

const systemPromptRewritePrefix = `You are an expert AI prompt engineer. Rewrite the provided prompt in the specified style while preserving its intent and meaning.

Rules:
- Preserve ALL original requirements and constraints — you may restructure but never drop information
- Maintain similar length to the original unless the style inherently requires more or less
- Respond in the same language as the input
- Do NOT add any explanations — return ONLY the rewritten prompt
- Do NOT wrap the result in quotes or code blocks` + antiPreamble + `

Style: `

var rewriteStyles = map[string]string{
	"formal":    "Formal — professional tone, business vocabulary, well-structured. Maintain similar length to the original.",
	"concise":   "Concise — compress and restructure into a dense, efficient form. Every requirement must be preserved but expressed more compactly. Merge related instructions, eliminate redundancy, use precise vocabulary. Result should be 30-50% shorter but feel more precise and powerful than the original, not just shorter.",
	"creative":  "Creative — rephrase the PROMPT ITSELF (the instructions) using vivid, inspiring language. The result must still be a prompt (instructions for an AI to follow), NOT the final content. Use an engaging framing, fresh wording, and an unexpected angle to make the instructions feel exciting. Maintain all original requirements.",
	"detailed":  "Detailed — expand with step-by-step instructions, expected input/output format, and edge cases. Generate only the detailed prompt itself, NOT an example of the expected output. Keep the total prompt under 1500 characters.",
	"technical": "Technical — rewrite as a formal specification or technical brief. Use numbered sections, precise terminology, input/output parameters, constraints, and acceptance criteria. Format like an engineering document.",
}

const systemPromptAnalyze = `You are an expert AI prompt quality evaluator. Analyze the provided prompt and give a rating.

Response format (strict, use these exact Russian headings):

## Оценка: X/10

## Сильные стороны
- ...

## Слабые стороны
- ...

## Рекомендации
1. ...
2. ...
3. ...

Rules:
- Evaluate: clarity, specificity, structure, context, constraints, output format
- Max 4 bullet points per section
- Recommendations must be actionable and specific — no generic advice
- Do NOT include example prompts in the analysis
- Be constructive and specific
- Respond in Russian` + antiPreamble

const systemPromptVariationsTemplate = `You are an expert AI prompt engineer. Generate %d different variations of the provided prompt. Each variation should address the same task but use a fundamentally different strategy.

Response format (strict, use these exact Russian headings):

## Вариант 1
[prompt text]

## Вариант 2
[prompt text]

## Вариант 3
[prompt text]

Rules:
- Each variation must use a fundamentally different STRATEGY: e.g., one role-based ("You are a..."), another constraint-driven ("Given these limitations..."), another goal-oriented ("Achieve X by..."). Do not just rephrase — change the approach.
- Respond in the same language as the input
- Do NOT add explanations to variations
- Use ONLY ## Вариант N headings as separators` + antiPreamble

func buildRewritePrompt(style RewriteStyle) string {
	desc, ok := rewriteStyles[string(style)]
	if !ok {
		desc = rewriteStyles["concise"]
	}
	return systemPromptRewritePrefix + desc
}

func buildVariationsPrompt(count int) string {
	if count <= 0 {
		count = 3
	}
	return fmt.Sprintf(systemPromptVariationsTemplate, count)
}
