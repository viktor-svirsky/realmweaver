import React, { useState } from 'react';
import { View, Text, Pressable, StyleSheet, Modal, ScrollView } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { NearbyPlayer, pvpChallenge, coopHelp, listTrades } from '../api/social';
import { rpc } from '../api/nakama';

interface Props {
  player: NearbyPlayer;
  onClose: () => void;
  onResult: (text: string) => void;
}

export default function PlayerInteraction({ player, onClose, onResult }: Props) {
  const character = useGameStore((s) => s.character);
  const [busy, setBusy] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [pvpRounds, setPvpRounds] = useState<Array<{ narrative: string }> | null>(null);

  if (!character) return null;

  async function handlePvP() {
    setBusy(true);
    try {
      const res = await pvpChallenge(character!.id, player.user_id, player.character_id);
      const data = res as {
        rounds: Array<{ narrative: string }>;
        winner: string;
        gold_reward: number;
      };
      setPvpRounds(data.rounds);
      setResult(`Winner: ${data.winner}! Gold reward: ${data.gold_reward}`);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'PvP failed';
      setResult(`Failed: ${msg}`);
    }
    setBusy(false);
  }

  async function handleHelp(type: 'heal' | 'buff_str' | 'buff_ac') {
    setBusy(true);
    try {
      const res = await coopHelp(
        character!.id,
        player.user_id,
        player.character_id,
        type,
      );
      setResult(res.result);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Help failed';
      setResult(`Failed: ${msg}`);
    }
    setBusy(false);
  }

  async function handleTradeGold() {
    setBusy(true);
    try {
      await rpc('trade_offer', {
        character_id: character!.id,
        offer_gold: 10,
        want_gold: 0,
        offer_item_id: '',
        want_item_id: '',
      });
      setResult('Trade offer posted: offering 10 gold');
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Trade failed';
      setResult(`Failed: ${msg}`);
    }
    setBusy(false);
  }

  const classIcon = player.character_class === 'warrior' ? '\u{1F6E1}\uFE0F' :
    player.character_class === 'mage' ? '\u{1F9D9}' :
    player.character_class === 'rogue' ? '\u{1F977}' : '\u{1F64F}';

  return (
    <Modal visible transparent animationType="slide">
      <View style={styles.overlay}>
        <View style={styles.sheet}>
          <View style={styles.header}>
            <View style={styles.playerInfo}>
              <Text style={styles.playerIcon}>{classIcon}</Text>
              <View>
                <Text style={styles.playerName}>{player.character_name}</Text>
                <Text style={styles.playerClass}>Lv{player.character_level} {player.character_class}</Text>
              </View>
            </View>
            <Pressable style={styles.closeBtn} onPress={onClose}>
              <Text style={styles.closeBtnText}>Close</Text>
            </Pressable>
          </View>

          <ScrollView contentContainerStyle={styles.content}>
            {/* PvP */}
            <Text style={styles.sectionTitle}>{'\u2694\uFE0F'} Combat</Text>
            <Pressable
              style={[styles.actionBtn, styles.pvpBtn, busy && styles.disabled]}
              onPress={handlePvP}
              disabled={busy}
            >
              <Text style={styles.actionBtnIcon}>{'\u{1F480}'}</Text>
              <View>
                <Text style={styles.actionBtnTitle}>Challenge to Duel</Text>
                <Text style={styles.actionBtnDesc}>5-round PvP. Loser pays 10% gold.</Text>
              </View>
            </Pressable>

            {/* Co-op */}
            <Text style={styles.sectionTitle}>{'\u{1F91D}'} Help</Text>
            <View style={styles.helpRow}>
              <Pressable
                style={[styles.helpBtn, busy && styles.disabled]}
                onPress={() => handleHelp('heal')}
                disabled={busy}
              >
                <Text style={styles.helpIcon}>{'\u2764\uFE0F'}</Text>
                <Text style={styles.helpLabel}>Heal</Text>
                <Text style={styles.helpCost}>5 MP</Text>
              </Pressable>
              <Pressable
                style={[styles.helpBtn, busy && styles.disabled]}
                onPress={() => handleHelp('buff_str')}
                disabled={busy}
              >
                <Text style={styles.helpIcon}>{'\u{1F4AA}'}</Text>
                <Text style={styles.helpLabel}>+2 STR</Text>
                <Text style={styles.helpCost}>3 MP</Text>
              </Pressable>
              <Pressable
                style={[styles.helpBtn, busy && styles.disabled]}
                onPress={() => handleHelp('buff_ac')}
                disabled={busy}
              >
                <Text style={styles.helpIcon}>{'\u{1F6E1}\uFE0F'}</Text>
                <Text style={styles.helpLabel}>+2 AC</Text>
                <Text style={styles.helpCost}>3 MP</Text>
              </Pressable>
            </View>

            {/* Trade */}
            <Text style={styles.sectionTitle}>{'\u{1FA99}'} Trade</Text>
            <Pressable
              style={[styles.actionBtn, styles.tradeBtn, busy && styles.disabled]}
              onPress={handleTradeGold}
              disabled={busy}
            >
              <Text style={styles.actionBtnIcon}>{'\u{1FA99}'}</Text>
              <View>
                <Text style={styles.actionBtnTitle}>Offer 10 Gold</Text>
                <Text style={styles.actionBtnDesc}>Post a gift for this player</Text>
              </View>
            </Pressable>

            {/* PvP Rounds */}
            {pvpRounds && (
              <View style={styles.pvpResults}>
                <Text style={styles.sectionTitle}>{'\u2694\uFE0F'} Duel Results</Text>
                {pvpRounds.map((round, i) => (
                  <Text key={i} style={styles.pvpRound}>{round.narrative}</Text>
                ))}
              </View>
            )}

            {/* Result */}
            {result && (
              <View style={styles.resultBox}>
                <Text style={styles.resultText}>{result}</Text>
              </View>
            )}
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.7)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: '#0d1117',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '80%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#1a1a2e',
  },
  playerInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  playerIcon: {
    fontSize: 36,
  },
  playerName: {
    color: '#7ecbf5',
    fontSize: 20,
    fontWeight: 'bold',
  },
  playerClass: {
    color: '#667',
    fontSize: 14,
  },
  closeBtn: {
    backgroundColor: '#2c3e50',
    paddingHorizontal: 20,
    paddingVertical: 8,
    borderRadius: 20,
  },
  closeBtnText: {
    color: '#e0d68a',
    fontWeight: 'bold',
    fontSize: 14,
  },
  content: {
    padding: 16,
    paddingBottom: 40,
  },
  sectionTitle: {
    color: '#e0d68a',
    fontSize: 16,
    fontWeight: 'bold',
    marginTop: 16,
    marginBottom: 10,
  },
  actionBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 14,
    borderRadius: 12,
    gap: 12,
    borderWidth: 1,
  },
  pvpBtn: {
    backgroundColor: '#2a1515',
    borderColor: '#c0392b',
  },
  tradeBtn: {
    backgroundColor: '#1a2a15',
    borderColor: '#27ae60',
  },
  disabled: {
    opacity: 0.4,
  },
  actionBtnIcon: {
    fontSize: 28,
  },
  actionBtnTitle: {
    color: '#d4d4d4',
    fontSize: 16,
    fontWeight: 'bold',
  },
  actionBtnDesc: {
    color: '#667',
    fontSize: 12,
    marginTop: 2,
  },
  helpRow: {
    flexDirection: 'row',
    gap: 10,
  },
  helpBtn: {
    flex: 1,
    backgroundColor: '#1a1a2e',
    borderRadius: 12,
    padding: 14,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2a2a3e',
  },
  helpIcon: {
    fontSize: 24,
  },
  helpLabel: {
    color: '#d4d4d4',
    fontSize: 13,
    fontWeight: 'bold',
    marginTop: 4,
  },
  helpCost: {
    color: '#2980b9',
    fontSize: 11,
    marginTop: 2,
  },
  pvpResults: {
    marginTop: 8,
  },
  pvpRound: {
    color: '#d4d4d4',
    fontSize: 13,
    marginBottom: 4,
    paddingLeft: 8,
    borderLeftWidth: 2,
    borderLeftColor: '#c0392b',
  },
  resultBox: {
    backgroundColor: 'rgba(224, 214, 138, 0.1)',
    padding: 12,
    borderRadius: 8,
    marginTop: 12,
  },
  resultText: {
    color: '#e0d68a',
    fontSize: 14,
    fontWeight: 'bold',
    textAlign: 'center',
  },
});
