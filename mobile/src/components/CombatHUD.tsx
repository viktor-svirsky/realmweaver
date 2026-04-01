import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useGameStore } from '../state/gameStore';
import { getNameColor, getPKTitle } from '../types/game';
import { t } from '../i18n/translations';

export default function CombatHUD() {
  const character = useGameStore((s) => s.character);
  const combat = useGameStore((s) => s.combat);
  const phase = useGameStore((s) => s.phase);
  const language = useGameStore((s) => s.language);

  if (phase !== 'in_combat' || !character) return null;

  const hpPct = Math.max(0, (character.hp / character.max_hp) * 100);
  const mpPct = Math.max(0, (character.mana / character.max_mana) * 100);

  const hpColor = hpPct > 50 ? '#27ae60' : hpPct > 25 ? '#f39c12' : '#c0392b';

  return (
    <View style={styles.container}>
      {/* Player */}
      <View style={styles.playerSection}>
        <View style={styles.playerNameRow}>
          <Text style={styles.playerName}>{character.name}</Text>
          <Text style={[styles.karmaTag, { color: getNameColor(character.karma, character.flagged) }]}>
            {getPKTitle(character.karma, character.pk_count)}
          </Text>
        </View>
        <View style={styles.barRow}>
          <Text style={styles.barLabel}>HP</Text>
          <View style={styles.barBg}>
            <View style={[styles.barFill, { width: `${hpPct}%`, backgroundColor: hpColor }]} />
          </View>
          <Text style={styles.barValue}>{character.hp}/{character.max_hp}</Text>
        </View>
        <View style={styles.barRow}>
          <Text style={styles.barLabel}>MP</Text>
          <View style={styles.barBg}>
            <View style={[styles.barFill, { width: `${mpPct}%`, backgroundColor: '#2980b9' }]} />
          </View>
          <Text style={styles.barValue}>{character.mana}/{character.max_mana}</Text>
        </View>
      </View>

      {/* Enemies */}
      {combat && combat.enemies && (
        <View style={styles.enemySection}>
          {combat.enemies.map((enemy) => {
            const ehpPct = Math.max(0, (enemy.hp / enemy.max_hp) * 100);
            const eColor = ehpPct > 50 ? '#c0392b' : ehpPct > 25 ? '#e74c3c' : '#922b21';
            const isDead = enemy.hp <= 0;
            return (
              <View key={enemy.id} style={[styles.enemyRow, isDead && styles.enemyDead]}>
                <Text style={[styles.enemyName, isDead && styles.enemyDeadText]}>
                  {isDead ? '\u{1F480}' : '\u{1F47F}'} {enemy.name}
                </Text>
                <View style={styles.enemyBarBg}>
                  <View style={[styles.barFill, { width: `${ehpPct}%`, backgroundColor: eColor }]} />
                </View>
                <Text style={styles.enemyHP}>{Math.max(0, enemy.hp)}/{enemy.max_hp}</Text>
              </View>
            );
          })}
        </View>
      )}

      {combat && (
        <Text style={styles.roundText}>{t('round', language)} {combat.round}</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    marginHorizontal: 12,
    borderRadius: 12,
    padding: 12,
    marginBottom: 8,
  },
  playerSection: {
    marginBottom: 8,
  },
  playerNameRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 4,
  },
  playerName: {
    color: '#e0d68a',
    fontSize: 14,
    fontWeight: 'bold',
  },
  karmaTag: {
    fontSize: 10,
    fontWeight: 'bold',
  },
  barRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 3,
    gap: 6,
  },
  barLabel: {
    color: '#888',
    fontSize: 10,
    fontWeight: 'bold',
    width: 20,
  },
  barBg: {
    flex: 1,
    height: 10,
    backgroundColor: '#1a1a1a',
    borderRadius: 5,
    overflow: 'hidden',
  },
  barFill: {
    height: '100%',
    borderRadius: 5,
  },
  barValue: {
    color: '#aaa',
    fontSize: 10,
    width: 45,
    textAlign: 'right',
  },
  enemySection: {
    borderTopWidth: 1,
    borderTopColor: '#ffffff15',
    paddingTop: 8,
    gap: 4,
  },
  enemyRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  enemyDead: {
    opacity: 0.4,
  },
  enemyName: {
    color: '#e74c3c',
    fontSize: 12,
    fontWeight: 'bold',
    width: 130,
  },
  enemyDeadText: {
    textDecorationLine: 'line-through',
    color: '#666',
  },
  enemyBarBg: {
    flex: 1,
    height: 8,
    backgroundColor: '#1a1a1a',
    borderRadius: 4,
    overflow: 'hidden',
  },
  enemyHP: {
    color: '#888',
    fontSize: 10,
    width: 40,
    textAlign: 'right',
  },
  roundText: {
    color: '#556',
    fontSize: 10,
    textAlign: 'center',
    marginTop: 6,
  },
});
