package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"realmweaver/engine"
	"realmweaver/world"
)

// GameContext assembles the full context string for a Claude API call.
type GameContext struct {
	Character    *engine.Character    `json:"character"`
	Location     *world.RegionSummary `json:"location"`
	Phase        engine.GamePhase     `json:"phase"`
	NPCContext   *NPCContext          `json:"npc_context,omitempty"`
	MechResult   *engine.ActionResult `json:"mechanical_result,omitempty"`
	RecentEvents []string             `json:"recent_events,omitempty"`
	PlayerAction string               `json:"player_action"`
	Language     string               `json:"language"`
}

// NPCContext holds NPC-specific info for dialogue.
type NPCContext struct {
	Name           string `json:"name"`
	Occupation     string `json:"occupation"`
	Personality    string `json:"personality"`
	Disposition    int    `json:"disposition"`
	LocationDetail string `json:"location_detail"`
}

// BuildUserMessage creates the user message for Claude from game context.
func BuildUserMessage(ctx *GameContext) string {
	var sb strings.Builder

	sb.WriteString("GAME STATE:\n")
	if ctx.Character != nil {
		sb.WriteString(fmt.Sprintf("Player: %s\n", ctx.Character.Summary()))
	}
	if ctx.Location != nil {
		locJSON, _ := json.Marshal(ctx.Location)
		sb.WriteString(fmt.Sprintf("Location: %s\n", string(locJSON)))
	}
	sb.WriteString(fmt.Sprintf("Phase: %s\n", ctx.Phase))

	if ctx.NPCContext != nil {
		sb.WriteString(fmt.Sprintf("\nNPC: %s (%s) — %s\nDisposition: %d\nPersonality: %s\n",
			ctx.NPCContext.Name, ctx.NPCContext.Occupation, ctx.NPCContext.LocationDetail,
			ctx.NPCContext.Disposition, ctx.NPCContext.Personality))
	}

	if ctx.MechResult != nil {
		mechJSON, _ := json.Marshal(ctx.MechResult)
		sb.WriteString(fmt.Sprintf("\nMECHANICAL RESULT:\n%s\n", string(mechJSON)))
	}

	if len(ctx.RecentEvents) > 0 {
		sb.WriteString("\nRECENT EVENTS:\n")
		for _, e := range ctx.RecentEvents {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	sb.WriteString(fmt.Sprintf("\nPLAYER ACTION: %s\n", ctx.PlayerAction))

	if ctx.Language != "" && ctx.Language != "en" {
		lang := languageName(ctx.Language)
		extra := ""
		if ctx.Language == "uk" {
			extra = " Use Ukrainian, NOT Russian."
		}
		sb.WriteString(fmt.Sprintf("\nIMPORTANT: Write the narrative value ENTIRELY in %s. Do NOT use English.%s\n", lang, extra))
	}
	sb.WriteString("\nNarrate this moment. Respond as JSON with narrative and hints.")

	return sb.String()
}

// SelectSystemPrompt picks the right system prompt based on game phase and language.
func SelectSystemPrompt(phase engine.GamePhase, language string) string {
	var base string

	if language != "" && language != "en" {
		lang := languageName(language)
		extra := ""
		if language == "uk" {
			extra = " You MUST use Ukrainian (\u0443\u043A\u0440\u0430\u0457\u043D\u0441\u044C\u043A\u0430), NOT Russian (\u0440\u0443\u0441\u0441\u043A\u0438\u0439). These are different languages."
		}
		base = fmt.Sprintf("CRITICAL INSTRUCTION: You MUST write ALL narrative text and dialogue EXCLUSIVELY in %s language. Do NOT use English for any narrative content. JSON keys stay in English but ALL text values MUST be in %s.%s\n\n", lang, lang, extra)
		base += SystemPrompt + "\n\n"
		base += fmt.Sprintf("REMINDER: Your narrative and all NPC dialogue MUST be written in %s.\n\n", lang)
	} else {
		base = SystemPrompt + "\n\n"
	}

	switch phase {
	case engine.PhaseInCombat:
		return base + CombatNarrationPrompt
	case engine.PhaseInDialogue:
		return base + DialoguePrompt
	default:
		return base + ExplorationPrompt
	}
}

func languageName(code string) string {
	names := map[string]string{
		"en": "English",
		"uk": "Ukrainian",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
		"pl": "Polish",
		"nl": "Dutch",
		"sv": "Swedish",
		"ru": "Russian",
		"ar": "Arabic",
		"tr": "Turkish",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}
