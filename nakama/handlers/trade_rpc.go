package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"

	"realmweaver/engine"
	"realmweaver/storage"
)

// RPCTradeOffer creates a trade offer visible to nearby players.
func RPCTradeOffer(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		CharacterID string `json:"character_id"`
		OfferItemID string `json:"offer_item_id"` // Item to give
		OfferGold   int    `json:"offer_gold"`     // Gold to give
		WantItemID  string `json:"want_item_id"`   // Item wanted (empty = any)
		WantGold    int    `json:"want_gold"`       // Gold wanted
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	character, err := storage.LoadCharacter(ctx, nk, userID, req.CharacterID)
	if err != nil || character == nil {
		return "", runtime.NewError("character not found", 5)
	}

	// Validate offer
	if req.OfferItemID != "" {
		item := engine.FindItem(character, req.OfferItemID)
		if item == nil {
			return "", runtime.NewError("you don't have that item", 3)
		}
	}
	if req.OfferGold > character.Gold {
		return "", runtime.NewError("not enough gold", 3)
	}

	// Store trade offer in Nakama storage (public read)
	tradeID := fmt.Sprintf("trade_%s_%d", userID[:8], engine.Roll(99999))
	offer := map[string]interface{}{
		"trade_id":       tradeID,
		"seller_id":      userID,
		"seller_name":    character.Name,
		"character_id":   req.CharacterID,
		"offer_item_id":  req.OfferItemID,
		"offer_gold":     req.OfferGold,
		"want_item_id":   req.WantItemID,
		"want_gold":      req.WantGold,
		"region_x":       character.RegionX,
		"region_y":       character.RegionY,
		"status":         "open",
	}

	// Add item name for display
	if req.OfferItemID != "" {
		item := engine.FindItem(character, req.OfferItemID)
		if item != nil {
			offer["offer_item_name"] = item.Name
		}
	}

	offerJSON, _ := json.Marshal(offer)
	ops := []*runtime.StorageWrite{
		{
			Collection:      "trades",
			Key:             tradeID,
			UserID:          userID,
			Value:           string(offerJSON),
			PermissionRead:  2, // public read
			PermissionWrite: 1, // owner write
		},
	}
	nk.StorageWrite(ctx, ops)

	data, _ := json.Marshal(offer)
	return string(data), nil
}

// RPCAcceptTrade accepts a trade offer from another player.
func RPCAcceptTrade(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	buyerID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || buyerID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		TradeID       string `json:"trade_id"`
		SellerID      string `json:"seller_id"`
		CharacterID   string `json:"character_id"` // Buyer's character
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	// Load trade offer with version for optimistic locking
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{Collection: "trades", Key: req.TradeID, UserID: req.SellerID},
	})
	if err != nil || len(objects) == 0 {
		return "", runtime.NewError("trade not found", 5)
	}

	tradeVersion := objects[0].Version // capture version for CAS

	var offer map[string]interface{}
	json.Unmarshal([]byte(objects[0].Value), &offer)

	if offer["status"] != "open" {
		return "", runtime.NewError("trade already completed", 3)
	}

	sellerID := offer["seller_id"].(string)
	sellerCharID := offer["character_id"].(string)

	// Load both characters
	buyerChar, err := storage.LoadCharacter(ctx, nk, buyerID, req.CharacterID)
	if err != nil || buyerChar == nil {
		return "", runtime.NewError("buyer character not found", 5)
	}

	sellerChar, err := storage.LoadCharacter(ctx, nk, sellerID, sellerCharID)
	if err != nil || sellerChar == nil {
		return "", runtime.NewError("seller character not found", 5)
	}

	// Validate BEFORE marking as completed (so trade stays open on failure)
	offerGold := int(offer["offer_gold"].(float64))
	wantGold := int(offer["want_gold"].(float64))

	if buyerChar.Gold < wantGold {
		return "", runtime.NewError("not enough gold", 3)
	}
	if sellerChar.Gold < offerGold {
		return "", runtime.NewError("seller no longer has enough gold", 3)
	}

	offerItemID, _ := offer["offer_item_id"].(string)
	if offerItemID != "" {
		item := engine.FindItem(sellerChar, offerItemID)
		if item == nil {
			return "", runtime.NewError("seller doesn't have the item anymore", 3)
		}
	}

	wantItemID, _ := offer["want_item_id"].(string)
	if wantItemID != "" {
		item := engine.FindItem(buyerChar, wantItemID)
		if item == nil {
			return "", runtime.NewError("you don't have the requested item", 3)
		}
	}

	// Atomically mark trade as completed using version check (prevents double-accept)
	offer["status"] = "completed"
	offer["buyer_id"] = buyerID
	offerJSON, _ := json.Marshal(offer)
	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{Collection: "trades", Key: req.TradeID, UserID: req.SellerID, Value: string(offerJSON), Version: tradeVersion, PermissionRead: 2, PermissionWrite: 1},
	})
	if err != nil {
		// Version conflict = another player already accepted
		return "", runtime.NewError("trade already accepted by another player", 3)
	}

	// Execute the trade (safe — only one accept can reach here, and we validated above)
	sellerChar.Gold -= offerGold
	buyerChar.Gold += offerGold
	buyerChar.Gold -= wantGold
	sellerChar.Gold += wantGold

	if offerItemID != "" {
		item, _ := engine.RemoveItem(sellerChar, offerItemID)
		if item != nil {
			engine.AddItem(buyerChar, *item)
		}
	}

	if wantItemID != "" {
		item, _ := engine.RemoveItem(buyerChar, wantItemID)
		if item != nil {
			engine.AddItem(sellerChar, *item)
		}
	}

	// Save both characters
	storage.SaveCharacter(ctx, nk, sellerID, sellerChar)
	storage.SaveCharacter(ctx, nk, buyerID, buyerChar)

	// Notify seller that their trade was accepted
	tradeContent := map[string]interface{}{
		"type":     "trade_accepted",
		"buyer":    buyerChar.Name,
		"trade_id": req.TradeID,
	}
	nk.NotificationSend(ctx, sellerID, "trade_accepted", tradeContent, 4, "", false)

	data, _ := json.Marshal(map[string]string{"status": "trade_complete"})
	return string(data), nil
}

// RPCListTrades returns open trade offers in the player's region.
func RPCListTrades(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", runtime.NewError("not authenticated", 16)
	}

	var req struct {
		RegionX int `json:"region_x"`
		RegionY int `json:"region_y"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "", runtime.NewError("invalid request", 3)
	}

	// List all public trades (we'll filter by region client-side for now)
	// In production, use a dedicated trades table with region indexing
	objects, _, err := nk.StorageList(ctx, "", "", "trades", 50, "")
	if err != nil {
		return "", runtime.NewError("failed to list trades", 13)
	}

	var trades []map[string]interface{}
	for _, obj := range objects {
		var trade map[string]interface{}
		json.Unmarshal([]byte(obj.Value), &trade)
		if trade["status"] == "open" {
			rx, _ := trade["region_x"].(float64)
			ry, _ := trade["region_y"].(float64)
			if int(rx) == req.RegionX && int(ry) == req.RegionY {
				trades = append(trades, trade)
			}
		}
	}

	data, _ := json.Marshal(map[string]interface{}{"trades": trades})
	return string(data), nil
}
